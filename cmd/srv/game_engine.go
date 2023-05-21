package main

import (
	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (app *App) startGameEngine(*cli.Context) error {
	app.loadStorage()
	app.loadRepos()
	app.loadPublisher()

	rooms, err := app.gameRepo.GetRooms(app.ctx)
	if err != nil {
		panic(err)
	}

	engineRouter := gameengine.NewRouter()
	requestSubscriber := kafka.NewSubscriber(
		"Engine",
		[]string{xcontext.Configs(app.ctx).Kafka.Addr},
		[]string{string(model.RequestTopic)},
		engineRouter.Subscribe,
	)

	for _, room := range rooms {
		_, err := gameengine.NewEngine(app.ctx, engineRouter, app.publisher, app.gameRepo, room.ID)
		if err != nil {
			panic(err)
		}
	}

	xcontext.Logger(app.ctx).Infof("Start game engine successfully")
	requestSubscriber.Subscribe(app.ctx)
	return nil
}
