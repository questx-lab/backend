package main

import (
	"fmt"
	"net/http"

	"github.com/urfave/cli/v2"
)

func (s *srv) startWsProxy(ctx *cli.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("ws", s.wsDomain.Serve)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.configs.WsProxyServer.Port),
		Handler: mux,
	}
	go s.wsDomain.Run()

	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
	fmt.Printf("server stop")
	return nil
}
