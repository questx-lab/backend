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
	projectRepo      repository.ProjectRepository
	claimedQuestRepo repository.ClaimedQuestRepository
}

func NewTrendingScoreCronJob(
	projectRepo repository.ProjectRepository,
	claimedQuestRepo repository.ClaimedQuestRepository,
) *TrendingScoreCronJob {
	return &TrendingScoreCronJob{
		projectRepo:      projectRepo,
		claimedQuestRepo: claimedQuestRepo,
	}
}

func (job *TrendingScoreCronJob) Do(ctx context.Context) {
	projects, err := job.projectRepo.GetList(ctx, repository.GetListProjectFilter{
		Offset: 0, Limit: -1,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get all projects: %v", err)
		return
	}

	startTime := dateutil.LastWeek(time.Now())
	endTime := startTime.AddDate(0, 0, 7)

	for _, p := range projects {
		trendingScore, err := job.claimedQuestRepo.Count(ctx, repository.CountClaimedQuestFilter{
			ProjectID:     p.ID,
			ReviewedStart: startTime,
			ReviewedEnd:   endTime,
			Status:        []entity.ClaimedQuestStatus{entity.Accepted, entity.AutoAccepted},
		})
		if err != nil {
			xcontext.Logger(ctx).Warnf("Cannot calculate trending score of project %s: %v", p.ID, err)
			continue
		}

		err = job.projectRepo.UpdateTrendingScore(ctx, p.ID, int(trendingScore))
		if err != nil {
			xcontext.Logger(ctx).Warnf("Cannot update trending score of project %s: %v", p.ID, err)
			continue
		}
	}
}

func (job *TrendingScoreCronJob) RunNow() bool {
	return false
}

func (job *TrendingScoreCronJob) Next() time.Time {
	return dateutil.NextWeek(time.Now())
}
