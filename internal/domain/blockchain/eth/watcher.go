package eth

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/blockchain/types"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"

	ethcommon "github.com/ethereum/go-ethereum/common"
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
	nftRepo        repository.NftRepository
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
	nftRepo repository.NftRepository,
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
		nftRepo:           nftRepo,
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
		xcontext.Logger(ctx).Debugf("%s block length = %d", w.chain, len(block.Transactions()))
		txs := w.processBlock(ctx, block)
		xcontext.Logger(ctx).Debugf("%s filtered txs = %d", w.chain, len(txs))

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
			Opts:        response.txs[i].Opts,
		}
	}
}

func (w *EthWatcher) processBlock(ctx context.Context, block *ethtypes.Block) []*types.TransactionWithOpts {
	ret := make([]*types.TransactionWithOpts, 0)

	for _, tx := range block.Transactions() {
		if ok, err := w.redisClient.Exist(ctx, tx.Hash().String()); ok && err == nil {
			opts, err := w.redisClient.Get(ctx, tx.Hash().String())
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot get redis tracked tx hash: %v", err)
				continue
			}

			if err := w.redisClient.Del(ctx, tx.Hash().String()); err != nil {
				xcontext.Logger(ctx).Warnf("Cannot delete redis tracked tx hash: %v", err)
			}

			ret = append(ret, &types.TransactionWithOpts{Transaction: tx, Opts: opts})
			continue
		}
	}

	return ret
}

func (w *EthWatcher) GetNonce(ctx context.Context, address string) (int64, error) {
	cAddr := ethcommon.HexToAddress(address)
	nonce, err := w.client.PendingNonceAt(ctx, cAddr)
	if err == nil {
		return int64(nonce), nil
	}

	return 0, fmt.Errorf("cannot get nonce of chain %s at %s", w.chain, address)
}

func (w *EthWatcher) TrackTx(ctx context.Context, txHash string) {
	xcontext.Logger(ctx).Infof("Tracking tx: %v", txHash)
	if err := w.redisClient.Set(ctx, txHash, ""); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to set txhash: %v", txHash)
	}
}

func (w *EthWatcher) TrackMintTx(ctx context.Context, txHash string, tokenID int64, amount int) {
	xcontext.Logger(ctx).Infof("Tracking tx: %v", txHash)
	if err := w.redisClient.Set(ctx, txHash, w.mintOpts(tokenID, amount)); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to set txhash: %v", txHash)
	}
}

func (w *EthWatcher) mintOpts(tokenID int64, amount int) string {
	return fmt.Sprintf("mint:%d:%d", tokenID, amount)
}

func (w *EthWatcher) parseMintOpts(opts string) (int64, int, error) {
	parts := strings.Split(opts, ":")
	if parts[0] != "mint" {
		return 0, 0, nil
	}

	if len(parts) != 3 {
		return 0, 0, errors.New("invalid parts of mint opts")
	}

	tokenID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	amount, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return tokenID, int(amount), nil
}

func (w *EthWatcher) updateTxs(ctx context.Context) {
	for {
		tx := <-w.txTrackCh

		_, err := w.blockChainRepo.GetTransactionByTxHash(ctx, tx.Hash.Hex(), tx.Chain)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Unable to retrieve tx_hash = %s, chain = %s", tx.Hash.String(), tx.Chain)
		}

		// step 1: confirm tx
		status := entity.BlockchainTransactionStatusTypeSuccess
		if tx.Result != types.TrackResultConfirmed {
			status = common.BlockchainTransactionFailure
		}

		if err := w.blockChainRepo.UpdateStatusByTxHash(ctx, tx.Hash.Hex(), tx.Chain, status); err != nil {
			xcontext.Logger(ctx).Errorf("Unable to update by txhash of tx_hash = %s, chain = %s", tx.Hash.String(), tx.Chain)
		}

		// Check if this is a mint tracking, update total balance.
		if tokenID, amount, err := w.parseMintOpts(tx.Opts); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot parse mint opts: %v", err)
		} else if tokenID != 0 && amount != 0 {
			if err := w.nftRepo.IncreaseTotalBalance(ctx, tokenID, amount); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot increase total balance: %v", err)
			}
		}
	}
}
