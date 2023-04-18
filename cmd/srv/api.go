package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/urfave/cli/v2"
)

func (s *srv) startApi(ct *cli.Context) error {
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.configs.ApiServer.Port),
		Handler: s.router.Handler(),
	}

	log.Printf("Starting server on port: %s\n", s.configs.ApiServer.Port)
	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
	log.Println("server stop")
	return nil
}
