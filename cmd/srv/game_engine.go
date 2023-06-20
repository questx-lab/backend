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

	engineRouter := gameengine.NewRouter(s.communityRepo, s.gameRepo, s.userRepo, s.storage, s.publisher)
	requestSubscriber := kafka.NewSubscriber(
		"Engine",
		[]string{xcontext.Configs(s.ctx).Kafka.Addr},
		[]string{model.GameActionRequestTopic, model.CreateCommunityTopic},
		engineRouter.Subscribe,
	)

	rooms, err := s.gameRepo.GetRoomsByCommunityID(s.ctx, "")
	if err != nil {
		panic(err)
	}

	for _, room := range rooms {
		_, err := gameengine.NewEngine(s.ctx, engineRouter, s.publisher,
			s.gameRepo, s.userRepo, s.storage, room.ID)
		if err != nil {
			panic(err)
		}
	}

	requestSubscriber.Subscribe(s.ctx)
	xcontext.Logger(s.ctx).Infof("Start game engine successfully")
	return nil
}
