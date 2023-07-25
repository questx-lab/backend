package main

import (
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/domain"
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGameProxy(*cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.migrateDB()
	s.loadStorage()
	s.loadRepos(nil)

	rpcNotificationEngineClient, err := rpc.DialContext(
		s.ctx, xcontext.Configs(s.ctx).Notification.EngineRPCServer.Endpoint)
	if err != nil {
		return err
	}

	proxyID := uuid.NewString()
	gameProxyDomain := domain.NewGameProxyDomain(proxyID, s.gameRepo, s.gameCharacterRepo,
		s.followerRepo, s.followerRoleRepo, s.userRepo, s.communityRepo,
		client.NewNotificationEngineCaller(rpcNotificationEngineClient))

	cfg := xcontext.Configs(s.ctx)
	defaultRouter := router.New(s.ctx)
	defaultRouter.AddCloser(middleware.Logger(cfg.Env))
	router.GET(defaultRouter, "/", homeHandle)

	authRouter := defaultRouter.Branch()
	authRouter.Before(middleware.NewAuthVerifier().WithAccessToken().Middleware())
	router.Websocket(authRouter, "/game", gameProxyDomain.ServeGameClient)

	xcontext.Logger(s.ctx).Infof("Server %s start in port: %s", proxyID, cfg.GameProxyServer.Port)

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
