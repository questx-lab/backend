package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/router"

	"github.com/urfave/cli/v2"
)

func (s *srv) startWsProxy(ctx *cli.Context) error {
	server.loadWsRouter()

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.configs.WsProxyServer.Port),
		Handler: s.router.Handler(),
	}

	// go routines for run websocket manager and consume kafka
	go s.wsDomain.Run()
	go s.responseSubscriber.Subscribe(context.Background())
	log.Printf("server start in port : %v\n", s.configs.WsProxyServer.Port)
	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
	log.Printf("server stop")
	return nil
}

func (s *srv) loadWsRouter() {
	s.router = router.New(s.db, *s.configs, s.logger)
	s.router.AddCloser(middleware.Logger())
	router.Websocket(s.router, "/test-game-client", s.wsDomain.ServeGameClient)
}
