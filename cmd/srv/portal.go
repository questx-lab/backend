package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/router"

	"github.com/urfave/cli/v2"
)

func (s *srv) startPortal(ct *cli.Context) error {
	server.loadConfig()
	server.loadLogger()
	server.loadEndpoint()
	server.loadDatabase()
	server.loadStorage()
	server.loadRepos()
	server.loadDomains()
	server.loadPortalRouter()

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.configs.PortalServer.Port),
		Handler: s.router.Handler(),
	}

	log.Printf("Starting server on port: %s\n", s.configs.PortalServer.Port)
	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
	log.Printf("server stop")
	return nil
}

func (s *srv) loadPortalRouter() {
	s.router = router.New(s.db, *s.configs, s.logger)
	s.router.AddCloser(middleware.Logger())

	authVerifier := middleware.NewAuthVerifier().WithAccessToken()

	s.router.Before(authVerifier.Middleware())

	router.POST(s.router, "/createMap", s.gameDomain.CreateMap)
	router.POST(s.router, "/createRoom", s.gameDomain.CreateRoom)
}
