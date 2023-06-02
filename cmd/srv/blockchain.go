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

	s.ethClients = map[string]eth.EthClient{
		"eth":              eth.NewEthClients(cfg.Eth.Chains["eth"], true),
		"ropsten-testnet":  eth.NewEthClients(cfg.Eth.Chains["ropsten-testnet"], true),
		"goerli-testnet":   eth.NewEthClients(cfg.Eth.Chains["goerli-testnet"], true),
		"xdai":             eth.NewEthClients(cfg.Eth.Chains["xdai"], true),
		"fantom-testnet":   eth.NewEthClients(cfg.Eth.Chains["fantom-testnet"], true),
		"polygon-testnet":  eth.NewEthClients(cfg.Eth.Chains["polygon-testnet"], true),
		"arbitrum-testnet": eth.NewEthClients(cfg.Eth.Chains["arbitrum-testnet"], true),
		"avaxc-testnet":    eth.NewEthClients(cfg.Eth.Chains["avaxc-testnet"], true),
	}

	return nil
}
