package main

import (
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/domain/notification/proxy"
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startNotificationProxy(*cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.migrateDB()
	s.loadRedisClient()
	s.loadRepos(nil)

	rpcNotificationEngineClient, err := rpc.DialContext(s.ctx,
		xcontext.Configs(s.ctx).Notification.EngineRPCServer.Endpoint)
	if err != nil {
		return err
	}

	notificationProxy := proxy.NewProxyServer(s.ctx, s.chatMemberRepo, s.chatChannelRepo,
		s.followerRepo, s.communityRepo, s.userRepo, s.redisClient,
		client.NewNotificationEngineCaller(rpcNotificationEngineClient))

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
