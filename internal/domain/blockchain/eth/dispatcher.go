package eth

import (
	"context"
	"math/big"
	"strings"

	"github.com/questx-lab/backend/internal/domain/blockchain/types"
	"github.com/questx-lab/backend/pkg/xcontext"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type EthDispatcher struct {
	client EthClient
}

func NewEhtDispatcher(client EthClient) *EthDispatcher {
	return &EthDispatcher{client: client}
}

func (d *EthDispatcher) Dispatch(ctx context.Context, request *types.DispatchedTxRequest) *types.DispatchedTxResult {
	from, err := ethtypes.Sender(ethtypes.NewEIP155Signer(request.Tx.ChainId()), request.Tx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get sender of transaction: %v", err)
		return types.NewDispatchTxError(request, types.ErrGeneric)
	}

	// Check the balance to see if we have enough native token.
	balance, err := d.client.BalanceAt(ctx, from, nil)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get balance for account %s: %v", from, err)
		return types.NewDispatchTxError(request, types.ErrGeneric)
	}

	minimum := new(big.Int).Mul(request.Tx.GasPrice(), big.NewInt(int64(request.Tx.Gas())))
	minimum = minimum.Add(minimum, request.Tx.Value())
	if minimum.Cmp(balance) == 1 {
		xcontext.Logger(ctx).Errorf(
			"Balance smaller than minimum required for this transaction, "+
				"from = %s, balance = %s, minimum = %s, chain = %s",
			from.String(), balance.String(), minimum.String(), request.Chain)
		return types.NewDispatchTxError(request, types.ErrNotEnoughBalance)
	}

	// Dispath tx.
	err = d.tryDispatchTx(ctx, request.Tx)
	if err == nil {
		xcontext.Logger(ctx).Infof("Tx is dispatched successfully for chain %s from %s txHash = %s",
			request.Chain, from, request.Tx.Hash())
		return types.NewDispatchTxSuccess(request)
	} else if strings.Contains(err.Error(), "already known") {
		// This is a tx submission duplication. It's possible that another node has submitted the same
		// transaction. This is counted as successful submission despite a returned error. Ethereum does
		// not return error code in its JSON RPC, so we have to rely on string matching.
		return types.NewDispatchTxSuccess(request)
	} else {
		xcontext.Logger(ctx).Errorf("Failed to dispatch tx: %v", err)
	}

	return types.NewDispatchTxError(request, types.ErrSubmitTx)
}

func (d *EthDispatcher) tryDispatchTx(ctx context.Context, tx *ethtypes.Transaction) error {
	return d.client.SendTransaction(ctx, tx)
}
