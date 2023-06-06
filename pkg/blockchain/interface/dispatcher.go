package interfaze

import (
	"context"

	"github.com/questx-lab/backend/pkg/blockchain/types"
)

// This is an interface for all dispatcher that sends transactions to different blockchain.
type Dispatcher interface {
	Start(ctx context.Context)
	Dispatch(ctx context.Context, request *types.DispatchedTxRequest) *types.DispatchedTxResult
}
