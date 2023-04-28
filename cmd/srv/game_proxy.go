package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/domain/gameproxy"
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/router"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGameProxy(ctx *cli.Context) error {
	server.loadConfig()
	server.loadLogger()
	server.loadDatabase()
	server.loadStorage()
	server.loadRepos()
	server.loadPublisher()
	server.loadGame()
	server.loadDomains()
	server.loadGameProxyRouter()

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.configs.GameProxyServer.Port),
		Handler: s.router.Handler(),
	}

	responseSubscriber := kafka.NewSubscriber(
		"proxy/"+uuid.NewString(),
		[]string{s.configs.Kafka.Addr},
		[]string{string(model.ResponseTopic)},
		s.proxyRouter.Subscribe,
	)

	go responseSubscriber.Subscribe(context.Background())

	log.Printf("server start in port : %v\n", s.configs.GameProxyServer.Port)
	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
	log.Printf("server stop")
	return nil
}

func (s *srv) loadGameProxyRouter() {
	s.router = router.New(s.db, *s.configs, s.logger)
	s.router.AddCloser(middleware.Logger())
	s.router.Before(middleware.NewAuthVerifier().WithAccessToken().Middleware())
	router.Websocket(s.router, "/game", s.gameProxyDomain.ServeGameClient)
}

func (s *srv) loadGame() {
	s.proxyRouter = gameproxy.NewRouter(s.logger)
}
