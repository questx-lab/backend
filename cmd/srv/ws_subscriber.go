package main

import (
	"context"
	"time"

	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/pubsub"

	"github.com/urfave/cli/v2"
)

func (s *srv) startWsSubscriber(ctx *cli.Context) error {
	s.subscriber = kafka.NewSubscriber(
		"subscriber",
		[]string{},
		[]string{},
		func(ctx context.Context, pack *pubsub.Pack, t time.Time) {
			s.publisher.Publish(ctx, "Response", pack)
		})
	return nil
}
