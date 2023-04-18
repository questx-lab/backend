package main

import (
	"context"
	"time"

	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/pubsub"

	"github.com/urfave/cli/v2"
)

func (s *srv) startSubscriber(ctx *cli.Context) error {
	s.subscriber = kafka.NewSubscriber(
		"subscriber",
		[]string{},
		[]string{},
		func(context.Context, *pubsub.Pack, time.Time) {})
	return nil
}
