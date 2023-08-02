package main

import (
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/domain/cron"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/urfave/cli/v2"
)

func (s *srv) startCron(*cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.migrateDB()
	s.loadRedisClient()
	s.loadRepos(nil)

	rpcNotificationEngineClient, err := rpc.DialContext(s.ctx,
		xcontext.Configs(s.ctx).Notification.EngineRPCServer.Endpoint)
	if err != nil {
		return err
	}

	cronJobManager := cron.NewCronJobManager()
	cronJobManager.Start(
		s.ctx,
		cron.NewTrendingScoreCronJob(s.communityRepo, s.claimedQuestRepo),
		cron.NewCleanupUserStatusCronJob(s.followerRepo, s.userRepo, s.redisClient,
			client.NewNotificationEngineCaller(rpcNotificationEngineClient)),
	)

	return nil
}
