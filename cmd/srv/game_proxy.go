package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/ws"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGameProxy(ctx *cli.Context) error {
	server.loadConfig()
	server.loadLogger()
	server.loadDatabase()
	server.loadStorage()
	server.loadRepos()
	server.loadPublisher()
	server.loadDomains()
	server.loadSubscriber()
	server.loadGameProxyRouter()
	server.loadGame()

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.configs.WsProxyServer.Port),
		Handler: s.router.Handler(),
	}

	go s.responseSubscriber.Subscribe(context.Background())
	go s.hub.Run()

	log.Printf("server start in port : %v\n", s.configs.WsProxyServer.Port)
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
	// router.Websocket(s.router, "/game", s.wsDomain.ServeGameClient)
	router.Websocket(s.router, "/game/v2", s.gameProxyDomain.ServeGameClientV2)
}

func (s *srv) loadGame() {
	s.hub = ws.NewHub()
}
