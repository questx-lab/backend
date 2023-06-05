package eth

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/questx-lab/backend/config"
	iface "github.com/questx-lab/backend/pkg/blockchain/interface"
	"github.com/questx-lab/backend/pkg/blockchain/types"
	"github.com/questx-lab/backend/pkg/ethutil"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type EthDispatcher struct {
	cfg    config.ChainConfig
	client EthClient
}

func NewEhtDispatcher(cfg config.ChainConfig, client EthClient) iface.Dispatcher {
	return &EthDispatcher{
		cfg:    cfg,
		client: client,
	}
}

// Start implements Dispatcher interface.
func (d *EthDispatcher) Start() {
	// Do nothing.
}

func (d *EthDispatcher) Dispatch(request *types.DispatchedTxRequest) *types.DispatchedTxResult {
	txBytes := request.Tx

	tx := &ethtypes.Transaction{}
	err := tx.UnmarshalBinary(txBytes)
	if err != nil {
		log.Println("Failed to unmarshal ETH transaction, err = ", err)
		return types.NewDispatchTxError(request, types.ErrMarshal)
	}

	from := ethutil.PublicKeyBytesToAddress(request.PubKey)
	// Check the balance to see if we have enough native token.
	balance, err := d.client.BalanceAt(context.Background(), from, nil)
	if balance == nil {
		log.Printf("Cannot get balance for account %s", from)
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
	err = d.tryDispatchTx(tx, request.Chain, from)
	if err == nil {
		log.Println("Tx is dispatched successfully for chain ", request.Chain, " from ", from,
			" txHash =", tx.Hash())
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
		log.Println("Failed to dispatch tx, err = ", err)
	}

	return types.NewDispatchTxError(request, types.ErrSubmitTx)
}

func (d *EthDispatcher) tryDispatchTx(tx *ethtypes.Transaction, chain string, from common.Address) error {
	return d.client.SendTransaction(context.Background(), tx)
}
