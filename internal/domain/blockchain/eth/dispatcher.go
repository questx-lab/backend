package eth

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/questx-lab/backend/internal/domain/blockchain/types"
	"github.com/questx-lab/backend/pkg/ethutil"
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
	txBytes := request.Tx

	tx := &ethtypes.Transaction{}
	err := tx.UnmarshalBinary(txBytes)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Failed to unmarshal ETH transaction: %v", err)
		return types.NewDispatchTxError(request, types.ErrMarshal)
	}

	from := ethutil.PublicKeyBytesToAddress(request.PubKey)
	// Check the balance to see if we have enough native token.
	balance, err := d.client.BalanceAt(ctx, from, nil)
	if balance == nil {
		xcontext.Logger(ctx).Errorf("Cannot get balance for account %s", from)
		return &types.DispatchedTxResult{
			Success: false,
			Chain:   request.Chain,
			TxHash:  request.TxHash,
			Err:     types.ErrGeneric,
		}
	}

	minimum := new(big.Int).Mul(tx.GasPrice(), big.NewInt(int64(tx.Gas())))
	minimum = minimum.Add(minimum, tx.Value())
	if minimum.Cmp(balance) > 0 {
		err = fmt.Errorf("balance smaller than minimum required for this transaction, from = %s, balance = %s, minimum = %s, chain = %s",
			from.String(), balance.String(), minimum.String(), request.Chain)
	}

	if err != nil {
		return &types.DispatchedTxResult{
			Success: false,
			Chain:   request.Chain,
			TxHash:  request.TxHash,
			Err:     types.ErrNotEnoughBalance,
		}
	}

	// Dispath tx.
	err = d.tryDispatchTx(ctx, tx)
	if err == nil {
		xcontext.Logger(ctx).Infof("Tx is dispatched successfully for chain %s from %s txHash = %s",
			request.Chain, from, tx.Hash())
		return &types.DispatchedTxResult{
			Success: true,
			Chain:   request.Chain,
			TxHash:  request.TxHash,
		}
	} else if strings.Contains(err.Error(), "already known") {
		// This is a tx submission duplication. It's possible that another node has submitted the same
		// transaction. This is counted as successful submission despite a returned error. Ethereum does
		// not return error code in its JSON RPC, so we have to rely on string matching.
		return &types.DispatchedTxResult{
			Success: true,
			Chain:   request.Chain,
			TxHash:  request.TxHash,
		}
	} else {
		xcontext.Logger(ctx).Errorf("Failed to dispatch tx: %v", err)
	}

	return types.NewDispatchTxError(request, types.ErrSubmitTx)
}

func (d *EthDispatcher) tryDispatchTx(ctx context.Context, tx *ethtypes.Transaction) error {
	return d.client.SendTransaction(ctx, tx)
}
