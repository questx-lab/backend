package main

import (
	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGameEngine(*cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.loadEndpoint()
	s.migrateDB()
	s.loadStorage()
	s.loadRepos()
	s.loadRedisClient()
	s.loadLeaderboard()
	s.loadPublisher()

	engineRouter := gameengine.NewRouter(
		s.communityRepo,
		s.gameRepo,
		s.gameLuckyboxRepo,
		s.gameCharacterRepo,
		s.userRepo,
		s.followerRepo,
		s.leaderboard,
		s.storage,
		s.publisher,
	)
	go engineRouter.PingCenter(s.ctx)

	subscriber := kafka.NewSubscriber(
		"engine/"+engineRouter.ID(),
		[]string{xcontext.Configs(s.ctx).Kafka.Addr},
		[]string{engineRouter.ID()},
		engineRouter.HandleEvent,
	)

	xcontext.Logger(s.ctx).Infof("Start game engine %s successfully", engineRouter.ID())
	subscriber.Subscribe(s.ctx)

	return nil
}
