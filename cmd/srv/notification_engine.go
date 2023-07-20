package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/internal/domain/notification/engine"
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startNotificationEngine(*cli.Context) error {
	cfg := xcontext.Configs(s.ctx)
	engineServer := engine.NewEngineServer()
	rpcHandler := rpc.NewServer()
	defer rpcHandler.Stop()

	err := rpcHandler.RegisterName(cfg.Notification.EngineRPCServer.RPCName, engineServer)
	if err != nil {
		return err
	}

	go func() {
		xcontext.Logger(s.ctx).Infof("Start rpc notification engine on port: %s",
			cfg.Notification.EngineRPCServer.Port)
		log.Println("Engine RPC address: ", cfg.Notification.EngineRPCServer.Address())
		httpSrv := &http.Server{
			Handler: rpcHandler,
			Addr:    fmt.Sprintf(":%v", cfg.Notification.EngineRPCServer.Port),
		}
		if err := httpSrv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	defaultRouter := router.New(s.ctx)
	router.Websocket(defaultRouter, "/", engineServer.ServeProxy)
	log.Println("Engine WS address: ", cfg.Notification.EngineWSServer.Address())
	defaultRouter.AddCloser(middleware.Logger(cfg.Env))
	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.Notification.EngineWSServer.Port),
		Handler: defaultRouter.Handler(cfg.Notification.EngineWSServer),
	}

	xcontext.Logger(s.ctx).Infof("Starting ws notification engine on port: %s",
		cfg.Notification.EngineWSServer.Port)
	if err := httpSrv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
