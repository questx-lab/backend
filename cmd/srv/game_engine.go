package main

import (
	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGameEngine(*cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.migrateDB()
	s.loadStorage()
	s.loadRepos()
	s.loadPublisher()

	rooms, err := s.gameRepo.GetRooms(s.ctx)
	if err != nil {
		panic(err)
	}

	engineRouter := gameengine.NewRouter()
	requestSubscriber := kafka.NewSubscriber(
		"Engine",
		[]string{xcontext.Configs(s.ctx).Kafka.Addr},
		[]string{string(model.RequestTopic)},
		engineRouter.Subscribe,
	)

	for _, room := range rooms {
		_, err := gameengine.NewEngine(s.ctx, engineRouter, s.publisher, s.gameRepo, room.ID)
		if err != nil {
			panic(err)
		}
	}

	xcontext.Logger(s.ctx).Infof("Start game engine successfully")
	requestSubscriber.Subscribe(s.ctx)
	return nil
}
