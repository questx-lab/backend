package main

import (
	"context"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGameProcessor(ctx *cli.Context) error {
	server.loadConfig()
	server.loadLogger()
	server.loadDatabase()
	server.loadStorage()
	server.loadRepos()
	server.loadPublisher()
	server.loadSubscriber()

	s.requestSubscriber.Subscribe(context.Background())
	return nil
}
