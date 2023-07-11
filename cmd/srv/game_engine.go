package main

import (
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/internal/domain/gameengine"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGameEngine(*cli.Context) error {
	cfg := xcontext.Configs(s.ctx)
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.loadEndpoint()
	s.migrateDB()
	s.loadStorage()
	s.loadRepos(nil)
	s.loadRedisClient()
	s.loadLeaderboard()

	centerRouterClient, err := rpc.DialContext(s.ctx, cfg.GameCenterServer.Endpoint)
	if err != nil {
		return err
	}

	engineRouter := gameengine.NewRouter(
		s.ctx,
		s.gameRepo,
		s.gameLuckyboxRepo,
		s.gameCharacterRepo,
		s.userRepo,
		s.followerRepo,
		s.leaderboard,
		s.storage,
		centerRouterClient,
	)
	go engineRouter.PingCenter(s.ctx, 0)

	rpcHandler := rpc.NewServer()
	defer rpcHandler.Stop()
	err = rpcHandler.RegisterName(cfg.GameEngineRPCServer.RPCName, engineRouter)
	if err != nil {
		return err
	}

	go func() {
		xcontext.Logger(s.ctx).Infof("Start rpc game engine %s successfully", engineRouter.ID())
		httpSrv := &http.Server{
			Handler: rpcHandler,
			Addr:    cfg.GameEngineRPCServer.Address(),
		}
		if err := httpSrv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	defaultRouter := router.New(s.ctx)
	router.Websocket(defaultRouter, "/proxy", engineRouter.ServeGameProxy)

	httpSrv := &http.Server{
		Addr:    cfg.GameEngineWSServer.Address(),
		Handler: defaultRouter.Handler(cfg.GameEngineWSServer),
	}
	xcontext.Logger(s.ctx).Infof("Starting ws game engine on port: %s", cfg.GameEngineWSServer.Port)
	if err := httpSrv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
