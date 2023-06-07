package eth

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	iface "github.com/questx-lab/backend/pkg/blockchain/interface"
	"github.com/questx-lab/backend/pkg/blockchain/types"
	"github.com/questx-lab/backend/pkg/ethutil"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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
	cfg              config.ChainConfig
	privKey          string
	client           EthClient
	blockTime        int
	blockChainTxRepo repository.BlockChainTransactionRepository
	vaultRepo        repository.VaultRepository
	txTrackCh        chan *types.TrackUpdate
	vaultAddress     string
	lock             *sync.RWMutex

	redisClient xredis.Client
	publisher   pubsub.Publisher

	// Block fetcher
	blockCh      chan *ethtypes.Block
	blockFetcher *defaultBlockFetcher

	// Receipt fetcher
	receiptFetcher    receiptFetcher
	receiptResponseCh chan *txReceiptResponse
}

func NewEthWatcher(
	vaultRepo repository.VaultRepository,
	blockChainTxRepo repository.BlockChainTransactionRepository,
	cfg config.ChainConfig,
	privKey string,
	client EthClient,
	redisClient xredis.Client,
	publisher pubsub.Publisher,
) iface.Watcher {
	blockCh := make(chan *ethtypes.Block)
	receiptResponseCh := make(chan *txReceiptResponse)

	w := &EthWatcher{
		privKey:           privKey,
		receiptResponseCh: receiptResponseCh,
		blockCh:           blockCh,
		blockFetcher:      newBlockFetcher(cfg, blockCh, client),
		receiptFetcher:    newReceiptFetcher(receiptResponseCh, client, cfg.Chain),
		blockChainTxRepo:  blockChainTxRepo,
		vaultRepo:         vaultRepo,
		cfg:               cfg,
		txTrackCh:         make(chan *types.TrackUpdate),
		blockTime:         cfg.BlockTime,
		client:            client,
		lock:              &sync.RWMutex{},
		redisClient:       redisClient,
		publisher:         publisher,
	}

	return w
}

func (w *EthWatcher) init(ctx context.Context) {
	privateKey, err := crypto.HexToECDSA(w.privKey)
	if err != nil {
		xcontext.Logger(ctx).Errorf("private key is not valid: %s", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		xcontext.Logger(ctx).Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	w.vaultAddress = address.String()

}

func (w *EthWatcher) Start(ctx context.Context) {
	xcontext.Logger(ctx).Infof("Starting Watcher...")

	w.init(ctx)
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
		xcontext.Logger(ctx).Infof(w.cfg.Chain, " Block length = ", len(block.Transactions()))
		txs := w.processBlock(ctx, block)
		xcontext.Logger(ctx).Infof(w.cfg.Chain, " Filtered txs = ", len(txs))

		if len(txs) > 0 {
			w.receiptFetcher.fetchReceipts(ctx, block.Number().Int64(), txs)
		}
	}
}

// waitForReceipt waits for receipts returned by the fetcher.
func (w *EthWatcher) waitForReceipt(ctx context.Context) {
	for {
		response := <-w.receiptResponseCh
		txs := w.extractTxs(ctx, response)

		xcontext.Logger(ctx).Infof(w.cfg.Chain, ": txs sizes = ", len(txs.Arr))

		// Save all txs into database for later references.
		if err := w.saveTxs(ctx, w.cfg.Chain, response.blockNumber, txs); err != nil {
			xcontext.Logger(ctx).Errorf("SaveTxs failed: ", err.Error())
		}
	}
}

func (w *EthWatcher) saveTxs(ctx context.Context, chain string, blockNumber int64, txs *types.Txs) error {
	for _, tx := range txs.Arr {
		hash := tx.Hash
		if len(hash) > 256 {
			hash = hash[:256]
		}
		err := w.blockChainTxRepo.CreateTransaction(ctx, &entity.BlockChainTransaction{
			Chain:       chain,
			TxHash:      hash,
			BlockHeight: blockNumber,
			TxBytes:     tx.Serialized,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// extractTxs takes response from the receipt fetcher and converts them into transactions.
func (w *EthWatcher) extractTxs(ctx context.Context, response *txReceiptResponse) *types.Txs {
	arr := make([]*types.Tx, 0)
	for i, tx := range response.txs {
		receipt := response.receipts[i]
		bz, err := tx.MarshalBinary()
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot serialize ETH tx, err = ", err)
			continue
		}

		if ok, err := w.redisClient.Exist(context.Background(), tx.Hash().String()); err == nil && ok {
			// Get Tx Receipt
			result := types.TrackResultConfirmed
			if receipt.Status == 0 {
				result = types.TrackResultFailure
			}

			// This is a transaction that we are tracking. Inform Sisu about this.
			w.txTrackCh <- &types.TrackUpdate{
				Chain:       w.cfg.Chain,
				Bytes:       bz,
				Hash:        tx.Hash(),
				BlockHeight: response.blockNumber,
				Result:      result,
			}

			continue
		}

		var to string
		if tx.To() == nil {
			to = ""
		} else {
			to = tx.To().String()
		}

		from, err := w.getFromAddress(w.cfg.Chain, tx)
		if err != nil {
			xcontext.Logger(ctx).Errorf("cannot get from address for tx %s on chain %s, err = %v\n", tx.Hash().String(), w.cfg.Chain, err)
			continue
		}

		arr = append(arr, &types.Tx{
			Hash:       tx.Hash().String(),
			Serialized: bz,
			From:       from.Hex(),
			To:         to,
			Success:    receipt.Status == 1,
		})
	}

	return &types.Txs{
		Chain:     w.cfg.Chain,
		Block:     response.blockNumber,
		BlockHash: response.blockHash,
		Arr:       arr,
	}
}

func (w *EthWatcher) processBlock(ctx context.Context, block *ethtypes.Block) []*ethtypes.Transaction {
	ret := make([]*ethtypes.Transaction, 0)

	for _, tx := range block.Transactions() {
		if ok, err := w.redisClient.Exist(ctx, tx.Hash().String()); ok && err == nil {
			ret = append(ret, tx)
			continue
		}

		if w.acceptTx(tx) {
			ret = append(ret, tx)
		}
	}

	return ret
}

func (w *EthWatcher) acceptTx(tx *ethtypes.Transaction) bool {
	if tx.To() != nil {
		if strings.EqualFold(tx.To().String(), w.vaultAddress) {
			return true
		}
	}

	return false
}

func (w *EthWatcher) getFromAddress(chain string, tx *ethtypes.Transaction) (common.Address, error) {
	signer := ethutil.GetEthChainSigner(chain)
	if signer == nil {
		return common.Address{}, fmt.Errorf("cannot find signer for chain %s", chain)
	}
	from, err := ethtypes.Sender(ethtypes.NewLondonSigner(tx.ChainId()), tx)
	if err != nil {
		from, err = ethtypes.Sender(ethtypes.HomesteadSigner{}, tx)
		if err != nil {
			return common.Address{}, err
		}
	}

	return from, nil
}

func (w *EthWatcher) GetNonce(ctx context.Context, address string) (int64, error) {
	cAddr := common.HexToAddress(address)
	nonce, err := w.client.PendingNonceAt(ctx, cAddr)
	if err == nil {
		return int64(nonce), nil
	}

	return 0, fmt.Errorf("cannot get nonce of chain %s at %s", w.cfg.Chain, address)
}

func (w *EthWatcher) TrackTx(ctx context.Context, txHash string) {
	xcontext.Logger(ctx).Infof("Tracking tx: ", txHash)
	w.redisClient.Set(context.Background(), txHash, txHash)
}

func (w *EthWatcher) updateTxs(ctx context.Context) {
	for {
		tx := <-w.txTrackCh
		// step 1: confirm tx
		if tx.Result != types.TrackResultConfirmed {
			receiptMsg := model.ReceiptMessage{
				TxHash:      tx.Hash.String(),
				BlockHeight: tx.BlockHeight,
				Timestamp:   time.Now(),
				TxStatus:    uint64(tx.Result),
			}

			w.publishTx(ctx, receiptMsg)
			continue
		}
		// step 2: fetch receipt (check tx successful or failed)
		ctx, cancel := context.WithTimeout(context.Background(), RpcTimeOut)
		receipt, err := w.client.TransactionReceipt(ctx, tx.Hash)
		cancel()

		if err != nil || receipt == nil {
			xcontext.Logger(ctx).Errorf("cannot get receipt for tx with hash %s on chain %s", tx.Hash.String(), tx.Chain)

			continue
		}

		receiptMsg := model.ReceiptMessage{
			ReceiptStatus: receipt.Status,
			TxHash:        tx.Hash.String(),
			BlockHeight:   tx.BlockHeight,
			Timestamp:     time.Now(),
			TxStatus:      uint64(tx.Result),
		}

		w.publishTx(ctx, receiptMsg)
	}
}

func (w *EthWatcher) publishTx(ctx context.Context, data model.ReceiptMessage) {

	b, err := json.Marshal(data)
	if err != nil {
		xcontext.Logger(ctx).Errorf("unable to marshal transaction = %v", data.TxHash)
		return
	}

	// step 3: update db and send message
	if err := w.publisher.Publish(ctx, model.ReceiptTransactionTopic, &pubsub.Pack{
		Key: []byte(uuid.NewString()),
		Msg: b,
	}); err != nil {
		xcontext.Logger(ctx).Errorf("unable to publish topic =  %v, transaction = %v", model.ReceiptTransactionTopic, data.TxHash)
	}
}
