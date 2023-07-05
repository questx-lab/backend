package domain

import (
	"context"
	"database/sql"
	"errors"

	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
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
	userID, communityID, invitedBy string,
) error {
	follower := &entity.Follower{
		UserID:      userID,
		CommunityID: communityID,
		InviteCode:  crypto.GenerateRandomAlphabet(9),
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	if invitedBy != "" {
		follower.InvitedBy = sql.NullString{String: invitedBy, Valid: true}
		err := followerRepo.IncreaseInviteCount(ctx, invitedBy, communityID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorx.New(errorx.NotFound, "Invalid invite user id")
			}

			xcontext.Logger(ctx).Errorf("Cannot increase invite: %v", err)
			return errorx.Unknown
		}

		err = badgeManager.
			WithBadges(badge.SharpScoutBadgeName).
			ScanAndGive(ctx, invitedBy, communityID)
		if err != nil {
			return err
		}
	}

	err := communityRepo.IncreaseFollowers(ctx, communityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot increase followers: %v", err)
		return errorx.Unknown
	}

	err = followerRepo.Create(ctx, follower)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create follower: %v", err)
		return errorx.Unknown
	}

	ctx = xcontext.WithCommitDBTransaction(ctx)

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

	return nil
}
