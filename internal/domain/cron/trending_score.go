package cron

import (
	"context"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/dateutil"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type TrendingScoreCronJob struct {
	communityRepo    repository.CommunityRepository
	claimedQuestRepo repository.ClaimedQuestRepository
}

func NewTrendingScoreCronJob(
	communityRepo repository.CommunityRepository,
	claimedQuestRepo repository.ClaimedQuestRepository,
) *TrendingScoreCronJob {
	return &TrendingScoreCronJob{
		communityRepo:    communityRepo,
		claimedQuestRepo: claimedQuestRepo,
	}
}

func (job *TrendingScoreCronJob) Do(ctx context.Context) {
	communities, err := job.communityRepo.GetList(ctx, repository.GetListCommunityFilter{})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all communities: %v", err)
		return
	}

	endTime := dateutil.BeginningOfDay(time.Now())
	startTime := endTime.AddDate(0, 0, -1)

	for _, p := range communities {
		trendingScore, err := job.claimedQuestRepo.Count(ctx, repository.StatisticClaimedQuestFilter{
			CommunityID:   p.ID,
			ReviewedStart: startTime,
			ReviewedEnd:   endTime,
			Status:        []entity.ClaimedQuestStatus{entity.Accepted, entity.AutoAccepted},
		})
		if err != nil {
			xcontext.Logger(ctx).Warnf("Cannot calculate trending score of community %s: %v", p.ID, err)
			continue
		}

		err = job.communityRepo.UpdateTrendingScore(ctx, p.ID, int(trendingScore))
		if err != nil {
			xcontext.Logger(ctx).Warnf("Cannot update trending score of community %s: %v", p.ID, err)
			continue
		}
	}
}

func (job *TrendingScoreCronJob) RunNow() bool {
	return true
}

func (job *TrendingScoreCronJob) Next() time.Time {
	return dateutil.NextDay(time.Now())
}
