package main

import (
	"fmt"
	"net/http"

	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/router"

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
	go s.wsDomain.Run()

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
