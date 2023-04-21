package interfaze

import "github.com/questx-lab/backend/pkg/blockchain/types"

// This is an interface for all dispatcher that sends transactions to different blockchain.
type Dispatcher interface {
	Start()
	Dispatch(request *types.DispatchedTxRequest) *types.DispatchedTxResult
}
