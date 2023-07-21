package domain

import (
	"context"
	"database/sql"
	"errors"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/domain/notification/event"
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
	followerRoleRepo repository.FollowerRoleRepository,
	badgeManager *badge.Manager,
	notificationEngineeCaller client.NotificationEngineCaller,
	userID, communityID, invitedBy string,
) error {
	follower := &entity.Follower{
		UserID:      userID,
		CommunityID: communityID,
		InviteCode:  crypto.GenerateRandomAlphabet(9),
	}

	followerRole := &entity.FollowerRole{
		UserID:      userID,
		CommunityID: communityID,
		RoleID:      entity.UserBaseRole,
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

	err = followerRoleRepo.Create(ctx, followerRole)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create follower role: %v", err)
		return errorx.Unknown
	}

	ctx = xcontext.WithCommitDBTransaction(ctx)

	community, err := communityRepo.GetByID(ctx, communityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return errorx.Unknown
	}

	go func() {
		ev := event.New(
			event.FollowCommunityEvent{
				CommunityID:     communityID,
				CommunityHandle: community.Handle,
			},
			&event.Metadata{ToUser: userID},
		)
		if err := notificationEngineeCaller.Emit(ctx, ev); err != nil {
			xcontext.Logger(ctx).Warnf("Cannot emit follow event: %v", err)
		}
	}()

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
