package types

import ethtypes "github.com/ethereum/go-ethereum/core/types"

type DispatchError int

const (
	ErrNil DispatchError = iota // no error
	ErrGeneric
	ErrNotEnoughBalance
	ErrMarshal
	ErrSubmitTx
	ErrNonceNotMatched
)

type DispatchedTxRequest struct {
	Chain string
	Tx    *ethtypes.Transaction
}

type DispatchedTxResult struct {
	Success bool
	Err     DispatchError // We use int since json RPC cannot marshal error
	Chain   string
	TxHash  string
}

func NewDispatchTxError(request *DispatchedTxRequest, err DispatchError) *DispatchedTxResult {
	return &DispatchedTxResult{
		Chain:   request.Chain,
		TxHash:  request.Tx.Hash().Hex(),
		Success: false,
		Err:     err,
	}
}

func NewDispatchTxSuccess(request *DispatchedTxRequest) *DispatchedTxResult {
	return &DispatchedTxResult{
		Chain:   request.Chain,
		TxHash:  request.Tx.Hash().Hex(),
		Success: true,
		Err:     ErrNil,
	}
}
