package main

import (
	"context"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGameProcessor(ctx *cli.Context) error {
	s.requestSubscriber.Subscribe(context.Background())
	return nil
}
