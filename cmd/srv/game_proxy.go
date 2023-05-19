package main

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/domain/gameproxy"
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGameProxy(*cli.Context) error {
	server.loadConfig()
	server.loadLogger()
	server.loadDatabase()
	server.loadStorage()
	server.loadRepos()
	server.loadPublisher()
	server.loadGame()
	server.loadDomains()
	server.loadGameProxyRouter()

	cfg := xcontext.Configs(s.ctx)
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.GameProxyServer.Port),
		Handler: s.router.Handler(cfg.GameProxyServer),
	}

	responseSubscriber := kafka.NewSubscriber(
		"proxy/"+uuid.NewString(),
		[]string{cfg.Kafka.Addr},
		[]string{string(model.ResponseTopic)},
		s.proxyRouter.Subscribe,
	)

	go responseSubscriber.Subscribe(s.ctx)

	xcontext.Logger(s.ctx).Infof("Server start in port : %v", cfg.GameProxyServer.Port)
	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
	xcontext.Logger(s.ctx).Infof("Server stop")

	return nil
}

func (s *srv) loadGameProxyRouter() {
	s.router = router.New(s.ctx)
	s.router.AddCloser(middleware.Logger())
	s.router.Before(middleware.NewAuthVerifier().WithAccessToken().Middleware())
	router.Websocket(s.router, "/game", s.gameProxyDomain.ServeGameClient)
}

func (s *srv) loadGame() {
	s.proxyRouter = gameproxy.NewRouter()
}
