package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/router"

	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
)

func (s *srv) startWsProxy(ctx *cli.Context) error {
	server.loadConfig()
	server.loadLogger()
	server.loadEndpoint()
	server.loadDatabase()
	server.loadStorage()
	server.loadRepos()
	server.loadDomains()
	server.loadWsRouter()

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.configs.WsProxyServer.Port),
		Handler: s.router.Handler(),
	}

	// for kafka flow
	kafkaAddr := s.configs.Kafka.Addr
	s.requestPublisher = kafka.NewPublisher(uuid.NewString(), []string{kafkaAddr})
	s.responseSubscriber = kafka.NewSubscriber(
		"subscriber",
		[]string{kafkaAddr},
		[]string{"RESPONSE"},
		s.wsDomain.WsSubscribeHandler,
	)

	// go routines for run websocket manager and consume kafka
	go s.wsDomain.Run()
	go s.responseSubscriber.Subscribe(context.Background())

	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
	fmt.Printf("server stop")
	return nil
}

func (s *srv) loadWsRouter() {
	s.router = router.New(s.db, *s.configs, s.logger)
	s.router.AddCloser(middleware.Logger())
	router.Websocket(s.router, "/test-game-client", s.wsDomain.ServeGameClient)
}
