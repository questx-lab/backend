package main

import (
	"context"
	"log"

	"github.com/questx-lab/backend/internal/domain/gameprocessor"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/kafka"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGameProcessor(ctx *cli.Context) error {
	server.loadConfig()
	server.loadLogger()
	server.loadDatabase()
	server.loadStorage()
	server.loadRepos()
	server.loadPublisher()

	requestSubscribeHandler := gameprocessor.NewRequestSubscribeHandler(s.publisher, s.logger)
	s.requestSubscriber = kafka.NewSubscriber(
		"Processor",
		[]string{s.configs.Kafka.Addr},
		[]string{string(model.RequestTopic)},
		requestSubscribeHandler.Subscribe,
	)

	log.Println("Start processor successful")
	s.requestSubscriber.Subscribe(context.Background())
	return nil
}
