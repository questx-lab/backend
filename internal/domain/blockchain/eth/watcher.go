package eth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/questx-lab/backend/internal/domain/blockchain/types"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const (
	minGasPrice      = 10_000_000_000
	TxTrackCacheSize = 1_000
)

type GasPriceGetter func(ctx context.Context) (*big.Int, error)

type BlockHeightExceededError struct {
	ChainHeight uint64
}

func NewBlockHeightExceededError(chainHeight uint64) error {
	return &BlockHeightExceededError{
		ChainHeight: chainHeight,
	}
}

func (e *BlockHeightExceededError) Error() string {
	return fmt.Sprintf("Our block height is higher than chain's height. Chain height = %d", e.ChainHeight)
}

type EthWatcher struct {
	chain          string
	client         EthClient
	blockChainRepo repository.BlockChainRepository
	txTrackCh      chan *types.TrackUpdate

	redisClient xredis.Client

	// Block fetcher
	blockCh      chan *ethtypes.Block
	blockFetcher *defaultBlockFetcher

	// Receipt fetcher
	receiptFetcher    receiptFetcher
	receiptResponseCh chan *txReceiptResponse
}

func NewEthWatcher(
	ctx context.Context,
	blockchain *entity.Blockchain,
	blockChainRepo repository.BlockChainRepository,
	client EthClient,
	redisClient xredis.Client,
) *EthWatcher {
	blockCh := make(chan *ethtypes.Block)
	receiptResponseCh := make(chan *txReceiptResponse)

	w := &EthWatcher{
		chain:             blockchain.Name,
		receiptResponseCh: receiptResponseCh,
		blockCh:           blockCh,
		blockFetcher:      newBlockFetcher(blockchain, blockCh, client),
		receiptFetcher:    newReceiptFetcher(receiptResponseCh, client, blockchain.Name),
		blockChainRepo:    blockChainRepo,
		txTrackCh:         make(chan *types.TrackUpdate),
		client:            client,
		redisClient:       redisClient,
	}

	return w
}

func (w *EthWatcher) Start(ctx context.Context) {
	xcontext.Logger(ctx).Infof("Starting Watcher...")

	go w.scanBlocks(ctx)
}

func (w *EthWatcher) scanBlocks(ctx context.Context) {
	go w.blockFetcher.start(ctx)
	go w.receiptFetcher.start(ctx)

	go w.waitForBlock(ctx)
	go w.waitForReceipt(ctx)
	go w.updateTxs(ctx)
}

// waitForBlock waits for new blocks from the block fetcher. It then filters interested txs and
// passes that to receipt fetcher to fetch receipt.
func (w *EthWatcher) waitForBlock(ctx context.Context) {
	for {
		block := <-w.blockCh

		// Pass this block to the receipt fetcher
		xcontext.Logger(ctx).Infof("%s block length = %d", w.chain, len(block.Transactions()))
		txs := w.processBlock(ctx, block)
		xcontext.Logger(ctx).Infof("%s filtered txs = %d", w.chain, len(txs))

		if len(txs) > 0 {
			w.receiptFetcher.fetchReceipts(ctx, block.Number().Int64(), txs)
		}
	}
}

// waitForReceipt waits for receipts returned by the fetcher.
func (w *EthWatcher) waitForReceipt(ctx context.Context) {
	for {
		response := <-w.receiptResponseCh
		w.extractTxs(ctx, response)
	}
}

// extractTxs takes response from the receipt fetcher and converts them into transactions.
func (w *EthWatcher) extractTxs(ctx context.Context, response *txReceiptResponse) {
	for i, tx := range response.txs {
		receipt := response.receipts[i]
		bz, err := tx.MarshalBinary()
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot serialize ETH tx, err = %v", err)
			continue
		}

		// Get Tx Receipt
		result := types.TrackResultConfirmed
		if receipt.Status == 0 {
			result = types.TrackResultFailure
		}

		// This is a transaction that we are tracking. Inform Sisu about this.
		w.txTrackCh <- &types.TrackUpdate{
			Chain:       w.chain,
			Bytes:       bz,
			Hash:        tx.Hash(),
			BlockHeight: response.blockNumber,
			Result:      result,
		}
	}
}

func (w *EthWatcher) processBlock(ctx context.Context, block *ethtypes.Block) []*ethtypes.Transaction {
	ret := make([]*ethtypes.Transaction, 0)

	for _, tx := range block.Transactions() {
		if ok, err := w.redisClient.Exist(ctx, tx.Hash().String()); ok && err == nil {
			if err := w.redisClient.Del(ctx, tx.Hash().String()); err != nil {
				xcontext.Logger(ctx).Warnf("Cannot delete redis tracked tx hash: %v", err)
			}

			ret = append(ret, tx)
			continue
		}
	}

	return ret
}

func (w *EthWatcher) GetNonce(ctx context.Context, address string) (int64, error) {
	cAddr := common.HexToAddress(address)
	nonce, err := w.client.PendingNonceAt(ctx, cAddr)
	if err == nil {
		return int64(nonce), nil
	}

	return 0, fmt.Errorf("cannot get nonce of chain %s at %s", w.chain, address)
}

func (w *EthWatcher) TrackTx(ctx context.Context, txHash string) {
	xcontext.Logger(ctx).Infof("Tracking tx: %v", txHash)
	if err := w.redisClient.Set(ctx, txHash, txHash); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to set txhash: %v", txHash)
	}
}

func (w *EthWatcher) updateTxs(ctx context.Context) {
	for {
		tx := <-w.txTrackCh

		_, err := w.blockChainRepo.GetTransactionByTxHash(ctx, tx.Hash.Hex(), tx.Chain)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Unable to retrieve tx_hash = %s, chain = %s", tx.Hash.String(), tx.Chain)
		}

		// step 1: confirm tx
		if tx.Result != types.TrackResultConfirmed {
			err := w.blockChainRepo.UpdateStatusByTxHash(
				ctx, tx.Hash.Hex(), tx.Chain, entity.BlockchainTransactionStatusTypeFailure)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Unable to update by txhash of tx_hash = %s, chain = %s", tx.Hash.String(), tx.Chain)
			}
			continue
		}

		// step 2: fetch receipt (check tx successful or failed)
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, RpcTimeOut)
		receipt, err := w.client.TransactionReceipt(ctx, tx.Hash)
		cancel()

		if err != nil || receipt == nil {
			xcontext.Logger(ctx).Errorf(
				"Cannot get receipt for tx with hash %s on chain %s", tx.Hash.String(), tx.Chain)
			continue
		}

		err = w.blockChainRepo.UpdateStatusByTxHash(
			ctx, tx.Hash.Hex(), tx.Chain, entity.BlockchainTransactionStatusTypeSuccess)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Unable to update by txhash of tx_hash = %s, chain = %s", tx.Hash.String(), tx.Chain)
		}
	}
}
