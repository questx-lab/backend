package interfaze

import "context"

type Watcher interface {
	Start(ctx context.Context)

	// Track a particular tx whose binary form on that chain is bz
	TrackTx(ctx context.Context, txHash string)
}
