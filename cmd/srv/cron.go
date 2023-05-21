package main

import (
	"github.com/questx-lab/backend/internal/domain/cron"
	"github.com/urfave/cli/v2"
)

func (app *App) startCron(*cli.Context) error {
	app.loadRepos()

	cronJobManager := cron.NewCronJobManager()
	cronJobManager.Register(cron.NewTrendingScoreCronJob(app.projectRepo, app.claimedQuestRepo))

	cronJobManager.Start(app.ctx)

	return nil
}
