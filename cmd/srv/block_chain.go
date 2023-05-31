package main

import (
	"github.com/questx-lab/backend/pkg/blockchain/eth"
	"github.com/questx-lab/backend/pkg/blockchain/types"
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
	txsCh := make(chan *types.Txs)
	txTrackCh := make(chan *types.TrackUpdate)
	watcher := eth.NewEthWatcher(s.vaultRepo, s.blockchainTxRepo, cfg.Chain, txsCh, txTrackCh, ethClient)

	dispatcher.Start()
	watcher.Start()

	return nil
}
