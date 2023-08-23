package main

import (
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/internal/domain/notification/engine"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startNotificationEngine(*cli.Context) error {
	engineServer := engine.NewEngineServer(s.ctx)
	rpcHandler := rpc.NewServer()
	defer rpcHandler.Stop()

	cfg := xcontext.Configs(s.ctx)
	err := rpcHandler.RegisterName(cfg.Notification.EngineRPCServer.RPCName, engineServer)
	if err != nil {
		return err
	}

	go func() {
		xcontext.Logger(s.ctx).Infof("Start rpc notification engine on port: %s",
			cfg.Notification.EngineRPCServer.Port)
		httpSrv := &http.Server{
			Handler: rpcHandler,
			Addr:    cfg.Notification.EngineRPCServer.Address(),
		}
		if err := httpSrv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	defaultRouter := router.New(s.ctx)
	router.Websocket(defaultRouter, "/proxy", engineServer.ServeProxy)
	httpSrv := &http.Server{
		Addr:    cfg.Notification.EngineWSServer.Address(),
		Handler: defaultRouter.Handler(cfg.Notification.EngineWSServer),
	}

	xcontext.Logger(s.ctx).Infof("Starting ws notification engine on port: %s",
		cfg.Notification.EngineWSServer.Port)
	if err := httpSrv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
