package main

import (
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/internal/domain/gamecenter"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startGameCenter(*cli.Context) error {
	cfg := xcontext.Configs(s.ctx)
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.migrateDB()
	s.loadStorage()
	s.loadRepos(nil)

	gameCenter := gamecenter.NewGameCenter(
		s.ctx,
		s.gameRepo,
		s.gameLuckyboxRepo,
		s.gameCharacterRepo,
		s.communityRepo,
		s.storage,
	)
	if err := gameCenter.Init(s.ctx); err != nil {
		return err
	}

	// Wait for some time to game center comsume all kafka events published
	// during downtime before calling load balance and janitor.
	time.AfterFunc(10*time.Second, func() {
		go gameCenter.Janitor(s.ctx)
		go gameCenter.LoadBalance(s.ctx)
		go gameCenter.ScheduleLuckyboxEvent(s.ctx)
	})

	rpcHandler := rpc.NewServer()
	defer rpcHandler.Stop()
	err := rpcHandler.RegisterName(cfg.GameCenterServer.RPCName, gameCenter)
	if err != nil {
		return err
	}

	xcontext.Logger(s.ctx).Infof("Start game center successfully")
	httpSrv := &http.Server{
		Handler: rpcHandler,
		Addr:    cfg.GameCenterServer.Address(),
	}
	if err := httpSrv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
