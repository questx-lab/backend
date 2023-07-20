package main

import (
	"net/http"

	"github.com/questx-lab/backend/internal/domain/notification/proxy"
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startNotificationProxy(*cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.migrateDB()
	s.loadRepos(nil)

	notificationProxy := proxy.NewProxyServer(s.ctx, s.chatMemberRepo, s.chatChannelRepo, s.followerRepo)

	cfg := xcontext.Configs(s.ctx)
	defaultRouter := router.New(s.ctx)
	defaultRouter.AddCloser(middleware.Logger(cfg.Env))
	defaultRouter.Before(middleware.NewAuthVerifier().WithAccessToken().Middleware())
	router.GET(defaultRouter, "/", homeHandle)
	router.Websocket(defaultRouter, "/notification", notificationProxy.ServeProxy)

	xcontext.Logger(s.ctx).Infof("Server start in port: %s", cfg.Notification.ProxyServer.Port)
	httpSrv := &http.Server{
		Addr:    cfg.Notification.ProxyServer.Address(),
		Handler: defaultRouter.Handler(cfg.Notification.ProxyServer),
	}
	if err := httpSrv.ListenAndServe(); err != nil {
		return err
	}

	xcontext.Logger(s.ctx).Infof("Server stop")
	return nil
}
