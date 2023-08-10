package cron

import (
	"context"
	"time"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/dateutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
)

type SetDailyCommunityStatCronJob struct {
	communityRepo repository.CommunityRepository
	userRepo      repository.UserRepository
	followerRepo  repository.FollowerRepository
	redisClient   xredis.Client
}

func NewSetDailyCommunityStatCronJob(
	communityRepo repository.CommunityRepository,
	userRepo repository.UserRepository,
	followerRepo repository.FollowerRepository,
	redisClient xredis.Client,
) *SetDailyCommunityStatCronJob {
	return &SetDailyCommunityStatCronJob{
		communityRepo: communityRepo,
		userRepo:      userRepo,
		followerRepo:  followerRepo,
		redisClient:   redisClient,
	}
}

func (job *SetDailyCommunityStatCronJob) Do(ctx context.Context) {
	communities, err := job.communityRepo.GetList(ctx, repository.GetListCommunityFilter{})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all communities to set stats: %v", err)
		return
	}

	for _, community := range communities {
		err := domain.CreateRedisFollowersIfNotExist(ctx, job.followerRepo, job.userRepo, job.redisClient, community.ID)
		if err != nil {
			continue
		}

		count, err := job.redisClient.SCard(ctx, common.RedisKeyFollower(community.ID))
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot count redis follower of %s: %v", community.ID, err)
			continue
		}

		err = job.communityRepo.SetStats(ctx, &entity.CommunityStats{
			CommunityID:   community.ID,
			Date:          dateutil.Date(time.Now()),
			FollowerCount: int(count),
		})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot set stats of community %s: %v", community.ID, err)
			continue
		}
	}
}

func (job *SetDailyCommunityStatCronJob) RunNow() bool {
	return true
}

func (job *SetDailyCommunityStatCronJob) Next() time.Time {
	return dateutil.NextDay(time.Now())
}
