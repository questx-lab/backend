package main

import (
	"context"
	"log"

	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGameEngine(cliCtx *cli.Context) error {
	server.loadConfig()
	server.loadLogger()
	server.loadDatabase()
	server.loadStorage()
	server.loadRepos()
	server.loadLogger()
	server.loadPublisher()

	ctx := context.Background()
	ctx = xcontext.WithConfigs(ctx, *s.configs)
	ctx = xcontext.WithLogger(ctx, s.logger)
	ctx = xcontext.WithDB(ctx, s.db)

	rooms, err := s.gameRepo.GetRooms(ctx)
	if err != nil {
		panic(err)
	}

	engineRouter := gameengine.NewRouter(s.logger)
	requestSubscriber := kafka.NewSubscriber(
		"Engine",
		[]string{s.configs.Kafka.Addr},
		[]string{string(model.RequestTopic)},
		engineRouter.Subscribe,
	)

	for _, room := range rooms {
		_, err := gameengine.NewEngine(ctx, engineRouter, s.publisher, s.logger, s.gameRepo, room.ID)
		if err != nil {
			panic(err)
		}
	}

	log.Println("Start game engine successfully")
	requestSubscriber.Subscribe(context.Background())
	return nil
}
