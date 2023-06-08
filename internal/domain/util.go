package domain

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

func followCommunity(
	ctx context.Context,
	userRepo repository.UserRepository,
	communityRepo repository.CommunityRepository,
	followerRepo repository.FollowerRepository,
	badgeManager *badge.Manager,
	userID, communityID, inviteCode string,
	explicitFollow bool,
) error {
	var inviteUser *entity.User
	if inviteCode != "" {
		var err error
		inviteUser, err = userRepo.GetByInviteCode(ctx, inviteCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorx.New(errorx.NotFound, "Not found user with invite code")
			}

			xcontext.Logger(ctx).Errorf("Cannot get invite user: %v", err)
			return errorx.Unknown
		}

		_, err = followerRepo.Get(ctx, inviteUser.ID, communityID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				xcontext.Logger(ctx).Errorf("Cannot get invite follower: %v", err)
				return errorx.Unknown
			}

			err = followCommunity(
				ctx,
				userRepo, communityRepo, followerRepo,
				badgeManager, inviteUser.ID, communityID,
				"",    // No invite user.
				false, // Implicitly follow (create a record then soft delete it).
			)
			if err != nil {
				return err
			}
		}
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	follower := &entity.Follower{
		UserID:      userID,
		CommunityID: communityID,
	}

	if !explicitFollow {
		// Soft delete the record if this is a implicit follow (not come from user request).
		follower.DeletedAt = gorm.DeletedAt{Valid: true, Time: time.Now()}
	}

	if inviteUser != nil {
		err := followerRepo.IncreaseInviteCount(ctx, inviteUser.ID, communityID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorx.New(errorx.NotFound, "Invalid invite user id")
			}

			xcontext.Logger(ctx).Errorf("Cannot increase invite: %v", err)
			return errorx.Unknown
		}

		err = badgeManager.
			WithBadges(badge.SharpScoutBadgeName).
			ScanAndGive(ctx, inviteUser.ID, communityID)
		if err != nil {
			return err
		}

		follower.InvitedBy = sql.NullString{String: inviteUser.ID, Valid: true}
	}

	err := followerRepo.Create(ctx, follower)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create follower: %v", err)
		return errorx.Unknown
	}

	err = communityRepo.IncreaseFollowers(ctx, communityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot increase followers: %v", err)
		return errorx.Unknown
	}

	community, err := communityRepo.GetByID(ctx, communityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return errorx.Unknown
	}

	isUnclaimable := community.ReferralStatus == entity.ReferralUnclaimable
	enoughFollowers := community.Followers >= xcontext.Configs(ctx).Quest.InviteCommunityRequiredFollowers
	if community.ReferredBy.Valid && enoughFollowers && isUnclaimable {
		err = communityRepo.UpdateByID(ctx, community.ID, entity.Community{
			ReferralStatus: entity.ReferralPending,
		})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot change referral status of community to pending: %v", err)
			return errorx.Unknown
		}
	}

	xcontext.WithCommitDBTransaction(ctx)
	return nil
}
