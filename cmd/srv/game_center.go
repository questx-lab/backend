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
	s.loadEndpoint()
	s.migrateDB()
	s.loadStorage()
	s.loadRepos()
	s.loadPublisher()

	gameCenter := gamecenter.NewGameCenter(
		s.gameRepo,
		s.gameCharacterRepo,
		s.communityRepo,
		s.publisher,
		s.storage,
	)
	if err := gameCenter.Init(s.ctx); err != nil {
		panic(err)
	}

	// Wait for some time to game center comsume all kafka events published
	// during downtime before calling load balance and janitor.
	time.AfterFunc(10*time.Second, func() {
		go gameCenter.Janitor(s.ctx)
		go gameCenter.LoadBalance(s.ctx)
	})

	subscriber := kafka.NewSubscriber(
		"GameCenter",
		[]string{xcontext.Configs(s.ctx).Kafka.Addr},
		[]string{model.CreateRoomTopic, model.GameEnginePingTopic},
		gameCenter.HandleEvent,
	)

	xcontext.Logger(s.ctx).Infof("Start game center successfully")
	subscriber.Subscribe(s.ctx)

	return nil
}
