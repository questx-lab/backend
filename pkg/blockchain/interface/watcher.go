package interfaze

import "context"

type Watcher interface {
	Start(ctx context.Context)

	// Set vault of the network. On chains like BTC, Cardano the gateway is the same as chain
	// account.
	SetVault(ctx context.Context, addr string, token string)

	// Track a particular tx whose binary form on that chain is bz
	TrackTx(ctx context.Context, txHash string)
}
