package cron

import (
	"context"
	"time"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/dateutil"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
)

type SetDailyCommunityRecordCronJob struct {
	communityRepo repository.CommunityRepository
	redisClient   xredis.Client
}

func NewSetDailyCommunityRecordCronJob(
	communityRepo repository.CommunityRepository,
	redisClient xredis.Client,
) *SetDailyCommunityRecordCronJob {
	return &SetDailyCommunityRecordCronJob{
		communityRepo: communityRepo,
		redisClient:   redisClient,
	}
}

func (job *SetDailyCommunityRecordCronJob) Do(ctx context.Context) {
	communities, err := job.communityRepo.GetList(ctx, repository.GetListCommunityFilter{})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all communities to set record: %v", err)
		return
	}

	for _, community := range communities {
		count, err := job.redisClient.SCard(ctx, common.RedisKeyFollower(community.ID))
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot count redis follower of %s: %v", community.ID, err)
			continue
		}

		err = job.communityRepo.SetRecord(ctx, &entity.CommunityRecord{
			CommunityID: community.ID,
			Date:        dateutil.Date(time.Now()),
			Followers:   int(count),
		})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot set record of community %s: %v", community.ID, err)
			continue
		}
	}
}

func (job *SetDailyCommunityRecordCronJob) RunNow() bool {
	return true
}

func (job *SetDailyCommunityRecordCronJob) Next() time.Time {
	return dateutil.NextDay(time.Now())
}
