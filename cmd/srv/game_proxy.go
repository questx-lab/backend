package main

import (
	"net/http"

	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGameProxy(*cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.loadEndpoint()
	s.migrateDB()
	s.loadStorage()
	s.loadRepos(nil)
	s.loadDomains(nil)

	cfg := xcontext.Configs(s.ctx)
	defaultRouter := router.New(s.ctx)
	defaultRouter.AddCloser(middleware.Logger(cfg.Env))
	router.GET(defaultRouter, "/", homeHandle)

	authRouter := defaultRouter.Branch()
	authRouter.Before(middleware.NewAuthVerifier().WithAccessToken().Middleware())
	router.Websocket(authRouter, "/game", s.gameProxyDomain.ServeGameClient)

	xcontext.Logger(s.ctx).Infof("Server start in port: %s", cfg.GameProxyServer.Port)

	httpSrv := &http.Server{
		Addr:    cfg.GameProxyServer.Address(),
		Handler: defaultRouter.Handler(cfg.GameProxyServer),
	}
	if err := httpSrv.ListenAndServe(); err != nil {
		return err
	}

	xcontext.Logger(s.ctx).Infof("Server stop")
	return nil
}
