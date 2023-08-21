package blockchain

import (
	"context"

	"github.com/questx-lab/backend/internal/domain/blockchain/types"
)

// This is an interface for all dispatcher that sends transactions to different blockchain.
type Dispatcher interface {
	Dispatch(ctx context.Context, request *types.DispatchedTxRequest) *types.DispatchedTxResult
}

type Watcher interface {
	Start(ctx context.Context)

	// Track a particular tx whose binary form on that chain is bz
	TrackTx(ctx context.Context, txHash string)
	TrackMintTx(ctx context.Context, txHash string, tokenID int64, amount int)
}
