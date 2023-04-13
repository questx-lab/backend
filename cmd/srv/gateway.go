package main

import (
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGateway(ct *cli.Context) error {
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.configs.Server.Port),
		Handler: s.router.Handler(),
	}

	fmt.Printf("Starting server on port: %s\n", s.configs.Server.Port)
	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
	fmt.Printf("server stop")
	return nil
}
