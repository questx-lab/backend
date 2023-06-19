package main

import (
	"time"

	"github.com/questx-lab/backend/internal/domain/gamecenter"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGameCenter(*cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.migrateDB()
	s.loadStorage()
	s.loadRepos()
	s.loadPublisher()

	gameCenter := gamecenter.NewGameCenter(
		s.gameRepo,
		s.communityRepo,
		s.publisher,
	)
	if err := gameCenter.Init(s.ctx); err != nil {
		panic(err)
	}

	time.AfterFunc(time.Minute, func() {
		go gameCenter.LoadBalance(s.ctx)
		go gameCenter.Janitor(s.ctx)
	})

	subscriber := kafka.NewSubscriber(
		"GameCenter",
		[]string{xcontext.Configs(s.ctx).Kafka.Addr},
		[]string{model.CreateCommunityTopic, model.GameEnginePingTopic},
		gameCenter.HandleEvent,
	)

	xcontext.Logger(s.ctx).Infof("Start game center successfully")
	subscriber.Subscribe(s.ctx)

	return nil
}
