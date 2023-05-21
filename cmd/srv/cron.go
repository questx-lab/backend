package main

import (
	"github.com/questx-lab/backend/internal/domain/cron"
	"github.com/urfave/cli/v2"
)

func (s *srv) startCron(*cli.Context) error {
	s.loadRepos()

	cronJobManager := cron.NewCronJobManager()
	cronJobManager.Register(cron.NewTrendingScoreCronJob(s.projectRepo, s.claimedQuestRepo))

	cronJobManager.Start(s.ctx)

	return nil
}
