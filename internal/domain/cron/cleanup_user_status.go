package cron

import (
	"context"
	"strconv"
	"time"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/notification/event"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
)

type CleanupUserStatusCronJob struct {
	followerRepo repository.FollowerRepository
	redisClient  xredis.Client
	engineCaller client.NotificationEngineCaller
}

func NewCleanupUserStatusCronJob(
	followerRepo repository.FollowerRepository,
	redisClient xredis.Client,
	engineCaller client.NotificationEngineCaller,
) *CleanupUserStatusCronJob {
	return &CleanupUserStatusCronJob{
		followerRepo: followerRepo,
		redisClient:  redisClient,
		engineCaller: engineCaller,
	}
}

func (job *CleanupUserStatusCronJob) Do(ctx context.Context) {
	userStatusKeys, err := job.redisClient.Keys(ctx, common.RedisKeyUserStatus("*"))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all user status keys: %v", err)
		return
	}

	if len(userStatusKeys) == 0 {
		return
	}

	pingTimes, err := job.redisClient.MGet(ctx, userStatusKeys...)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all ping time of user status: %v", err)
		return
	}

	now := time.Now().Unix()
	offlineUserStatusKeys := []string{}
	offlineUserIDs := []string{}
	for i := range userStatusKeys {
		if pingTimes[i] == nil {
			xcontext.Logger(ctx).Warnf("No value at key %s", userStatusKeys[i])
			continue
		}

		lastPingString, ok := pingTimes[i].(string)
		if !ok {
			xcontext.Logger(ctx).Errorf("Invalid type of ping time: %T", pingTimes[i])
			continue
		}

		lastPing, err := strconv.ParseInt(lastPingString, 10, 64)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot model.Convert ping value to int64: %v", err)
			continue
		}

		if now-lastPing > 30 { // 30 seconds
			userID := common.FromRedisKeyUserStatus(userStatusKeys[i])
			followers, err := job.followerRepo.GetListByUserID(ctx, userID)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot get followers: %v", err)
				continue
			}

			communityIDs := []string{}
			for _, f := range followers {
				communityIDs = append(communityIDs, f.CommunityID)
			}

			ev := event.New(
				event.ChangeUserStatusEvent{UserID: userID, Status: event.Offline},
				&event.Metadata{ToCommunities: communityIDs},
			)
			if err := job.engineCaller.Emit(ctx, ev); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot emit offline event: %v", err)
				continue
			}

			offlineUserStatusKeys = append(offlineUserStatusKeys, userStatusKeys[i])
			offlineUserIDs = append(offlineUserIDs, userID)
		}
	}

	if len(offlineUserIDs) == 0 {
		return
	}

	communityKeys, err := job.redisClient.Keys(ctx, common.RedisKeyCommunityOnline("*"))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all community online keys: %v", err)
		return
	}

	for _, communityKey := range communityKeys {
		if err := job.redisClient.SRem(ctx, communityKey, offlineUserIDs...); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot remove community online: %v", err)
			continue
		}
	}

	if err := job.redisClient.Del(ctx, offlineUserStatusKeys...); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot delete offline user keys: %v", err)
	}
}

func (job *CleanupUserStatusCronJob) RunNow() bool {
	return true
}

func (job *CleanupUserStatusCronJob) Next() time.Time {
	return time.Now().Add(time.Minute)
}
