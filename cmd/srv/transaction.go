package main

import (
	"github.com/questx-lab/backend/internal/domain"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/blockchain/eth"
	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/urfave/cli/v2"
)

func (s *srv) startTransaction(*cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.migrateDB()
	s.loadRepos()

	cfg := xcontext.Configs(s.ctx)
	ethClient := eth.NewEthClients(cfg.Chain, false)

	dispatcher := eth.NewEhtDispatcher(cfg.Chain, ethClient)
	watcher := eth.NewEthWatcher(
		s.vaultRepo,
		s.blockchainTxRepo,
		cfg.Chain,
		ethClient,
		s.redisClient,
		s.publisher,
	)
	s.dispatcherDomain = domain.NewDispatcherDomain(cfg.Chain, dispatcher, watcher, ethClient, s.transactionRepo)

	dispatchSubscriber := kafka.NewSubscriber(
		"dispatcher",
		[]string{cfg.Kafka.Addr},
		[]string{string(model.ReceiptTransactionTopic)},
		s.dispatcherDomain.Subscribe,
	)

	go watcher.Start()
	go dispatchSubscriber.Subscribe(s.ctx)

	return nil
}
