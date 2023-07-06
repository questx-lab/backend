package main

import (
	"github.com/questx-lab/backend/internal/domain/cron"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/urfave/cli/v2"
)

func (s *srv) startCron(*cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.loadEndpoint()
	s.migrateDB()
	s.loadRepos(nil)
	s.loadPublisher()

	cronJobManager := cron.NewCronJobManager()
	cronJobManager.Start(
		s.ctx,
		cron.NewTrendingScoreCronJob(s.communityRepo, s.claimedQuestRepo),
		cron.NewLuckyboxEventCronJob(s.gameRepo, s.publisher),
	)

	return nil
}
