package domain

import (
	"context"
	"database/sql"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/notification/event"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
)

func FollowCommunity(
	ctx context.Context,
	userRepo repository.UserRepository,
	communityRepo repository.CommunityRepository,
	followerRepo repository.FollowerRepository,
	followerRoleRepo repository.FollowerRoleRepository,
	notificationEngineeCaller client.NotificationEngineCaller,
	redisClient xredis.Client,
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

	followerKey := common.RedisKeyFollower(communityID)
	if exist, err := redisClient.Exist(ctx, followerKey); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot check existence of follower key: %v", err)
	} else if exist {
		user, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get user when follow: %v", err)
		} else {
			err := redisClient.SAdd(ctx, followerKey, common.RedisValueFollower(user.Name, user.ID))
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot add user to redis: %v", err)
			}
		}
	}

	go func() {
		if notificationEngineeCaller == nil {
			xcontext.Logger(ctx).Errorf("Cannot emit follow event: not found caller")
			return
		}

		ev := event.New(
			event.FollowCommunityEvent{
				CommunityID:     communityID,
				CommunityHandle: community.Handle,
			},
			&event.Metadata{ToUsers: []string{userID}},
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

func CreateRedisFollowersIfNotExist(
	ctx context.Context,
	followerRepo repository.FollowerRepository,
	userRepo repository.UserRepository,
	redisClient xredis.Client,
	communityID string,
) error {
	key := common.RedisKeyFollower(communityID)
	if exist, err := redisClient.Exist(ctx, key); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot check existence of follower key: %v", err)
		return errorx.Unknown
	} else if !exist {
		followers, err := followerRepo.GetListByCommunityID(ctx, repository.GetListFollowerFilter{
			CommunityID:    communityID,
			IgnoreUserRole: false,
			Offset:         0,
			Limit:          -1,
		})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get followers: %v", err)
			return errorx.Unknown
		}

		userIDs := []string{}
		for _, f := range followers {
			userIDs = append(userIDs, f.UserID)
		}

		users, err := userRepo.GetByIDs(ctx, userIDs)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get users: %v", err)
			return errorx.Unknown
		}

		usernames := []string{}
		for _, u := range users {
			usernames = append(usernames, common.RedisValueFollower(u.Name, u.ID))
		}

		if err := redisClient.SAdd(ctx, key, usernames...); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot add to redis: %v", err)
			return errorx.Unknown
		}
	}

	return nil
}
