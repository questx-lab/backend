package main

import (
	"github.com/questx-lab/backend/pkg/blockchain/eth"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startBlockchain(*cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.migrateDB()
	s.loadRepos()

	cfg := xcontext.Configs(s.ctx)
	ethClient := eth.NewEthClients(cfg.Chain, false)
	dispatcher := eth.NewEhtDispatcher(cfg.Chain, ethClient)
	// watcher := eth.NewEthWatcher(s.vaultRepo,s.blockchainTxRepo,cfg.Chain)

	dispatcher.Start()

	return nil
}
