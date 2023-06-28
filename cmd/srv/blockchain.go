package main

import (
	"github.com/puzpuzpuz/xsync"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/blockchain/eth"
	interfaze "github.com/questx-lab/backend/pkg/blockchain/interface"
	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/urfave/cli/v2"
)

func (s *srv) startBlockchain(*cli.Context) error {
	s.ctx = xcontext.WithDB(s.ctx, s.newDatabase())
	s.loadEndpoint()
	s.migrateDB()
	s.loadRepos()
	s.loadEthClients()
	s.loadDomains()

	return nil
}

func (s *srv) loadEthClients() {
	cfg := xcontext.Configs(s.ctx)

	ethChains := []string{"eth", "ropsten-testnet", "goerli-testnet", "xdai", "fantom-testnet", "polygon-testnet", "arbitrum-testnet", "avaxc-testnet"}
	s.ethClients = xsync.NewMapOf[eth.EthClient]()
	s.watchers = xsync.NewMapOf[interfaze.Watcher]()
	s.dispatchers = xsync.NewMapOf[interfaze.Dispatcher]()
	for _, chain := range ethChains {
		client := eth.NewEthClients(cfg.Eth.Chains[chain], true)
		dispatcher := eth.NewEhtDispatcher(cfg.Eth.Chains[chain], client)
		watcher := eth.NewEthWatcher(
			s.blockchainTransactionRepo,
			cfg.Eth.Chains[chain],
			cfg.Eth.Keys.PrivKey,
			client,
			s.redisClient,
			s.publisher,
		)
		s.ethClients.Store(chain, client)
		s.dispatchers.Store(chain, dispatcher)
		s.watchers.Store(chain, watcher)

		go dispatcher.Start(s.ctx)
		go watcher.Start(s.ctx)
	}

	payRewardSubscriber := kafka.NewSubscriber(
		"blockchain",
		[]string{cfg.Kafka.Addr},
		[]string{string(model.CreateTransactionTopic)},
		s.payRewardDomain.Subscribe,
	)

	go payRewardSubscriber.Subscribe(s.ctx)

}
