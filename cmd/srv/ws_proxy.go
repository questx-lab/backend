package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startWsProxy(ctx *cli.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("ws", func(w http.ResponseWriter, r *http.Request) {
		ctx := xcontext.NewContext(r.Context(), r, w, *s.configs, s.logger, s.db)
		if err := s.wsDomain.Serve(ctx); err != nil {
			resp := router.NewErrorResponse(err)
			if err := router.WriteJson(w, resp); err != nil {
				log.Println("unable to write json")
			}
		}
	})

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.configs.WsProxyServer.Port),
		Handler: mux,
	}
	go s.wsDomain.Run()

	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
	log.Println("server stop")
	return nil
}
