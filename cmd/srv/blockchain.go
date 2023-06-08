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
	s.loadEthClients()
	s.loadDomains()

	return nil
}

func (s *srv) loadEthClients() {
	cfg := xcontext.Configs(s.ctx)

	ethChains := []string{"eth", "ropsten-testnet", "goerli-testnet", "xdai", "fantom-testnet", "polygon-testnet", "arbitrum-testnet", "avaxc-testnet"}

	for _, chain := range ethChains {
		s.ethClients[chain] = eth.NewEthClients(cfg.Eth.Chains[chain], true)
		s.dispatchers[chain] = eth.NewEhtDispatcher(cfg.Eth.Chains[chain], s.ethClients[chain])
		s.watchers[chain] = eth.NewEthWatcher(
			s.blockchainTransactionRepo,
			cfg.Eth.Chains[chain],
			cfg.Eth.Keys.PrivKey,
			s.ethClients[chain],
			s.redisClient,
			s.publisher,
		)
	}

}
