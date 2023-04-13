package main

import (
	"github.com/questx-lab/backend/pkg/kafka"

	"github.com/urfave/cli/v2"
)

func (s *srv) startSubscriber(ctx *cli.Context) error {
	s.subscriber = kafka.NewSubscriber("subscriber", []string{}, []string{})
	return nil
}
