package main

import (
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/internal/domain/blockchain"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/urfave/cli/v2"
)

func (s *srv) startBlockchain(*cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.migrateDB()
	s.loadRedisClient()
	s.loadRepos(nil)

	blockchainManager := blockchain.NewBlockchainManager(
		s.ctx,
		s.payRewardRepo,
		s.communityRepo,
		s.blockchainRepo,
		s.redisClient,
	)

	go blockchainManager.Run(s.ctx)

	rpcHandler := rpc.NewServer()
	defer rpcHandler.Stop()
	err := rpcHandler.RegisterName(xcontext.Configs(s.ctx).Blockchain.RPCName, blockchainManager)
	if err != nil {
		xcontext.Logger(s.ctx).Infof("Cannot register blockchain manager: %v", err)
		return err
	}

	xcontext.Logger(s.ctx).Infof("Started rpc server of block chain")
	httpSrv := &http.Server{
		Handler: rpcHandler,
		Addr:    xcontext.Configs(s.ctx).Blockchain.Address(),
	}

	if err := httpSrv.ListenAndServe(); err != nil {
		xcontext.Logger(s.ctx).Errorf("An error occurs when running rpc server: %v", err)
		return err
	}

	return nil
}
