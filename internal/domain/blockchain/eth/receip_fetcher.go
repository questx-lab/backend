package eth

import (
	"context"
	"time"

	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/questx-lab/backend/internal/domain/blockchain/types"
	"github.com/questx-lab/backend/pkg/xcontext"
)

const (
	MaxReceiptRetry = 5
)

// txReceiptRequest is a data structure for this watcher to send request to the receipt fetcher.
type txReceiptRequest struct {
	blockNumber int64
	blockHash   string
	txs         []*types.TransactionWithOpts
}

// txReceiptResponse is a data structure for the receipt fetcher to return its result
type txReceiptResponse struct {
	blockNumber int64
	blockHash   string
	txs         []*types.TransactionWithOpts
	receipts    []*etypes.Receipt
}

type receiptFetcher interface {
	start(ctx context.Context)
	fetchReceipts(ctx context.Context, block int64, txs []*types.TransactionWithOpts)
}

type defaultReceiptFetcher struct {
	chain      string
	requestCh  chan *txReceiptRequest
	responseCh chan *txReceiptResponse
	retryTime  time.Duration

	client EthClient
}

func newReceiptFetcher(responseCh chan *txReceiptResponse, client EthClient, chain string) receiptFetcher {
	return &defaultReceiptFetcher{
		chain:      chain,
		requestCh:  make(chan *txReceiptRequest, 20),
		responseCh: responseCh,
		client:     client,
		retryTime:  time.Second * 5,
	}
}

func (rf *defaultReceiptFetcher) start(ctx context.Context) {
	for {
		request := <-rf.requestCh
		response := rf.getResponse(ctx, request)

		// Post the response
		rf.responseCh <- response
	}
}

func (rf *defaultReceiptFetcher) getResponse(ctx context.Context, request *txReceiptRequest) *txReceiptResponse {
	retry := 0
	response := &txReceiptResponse{
		blockNumber: request.blockNumber,
		blockHash:   request.blockHash,
		txs:         make([]*types.TransactionWithOpts, 0),
		receipts:    make([]*etypes.Receipt, 0),
	}

	txQueue := request.txs

	for {
		if len(txQueue) == 0 {
			break
		}

		ok := false
		tx := txQueue[0]

		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, RpcTimeOut)
		receipt, err := rf.client.TransactionReceipt(ctx, tx.Hash())
		cancel()

		if err == nil && receipt != nil {
			ok = true
			response.txs = append(response.txs, tx)
			response.receipts = append(response.receipts, receipt)
		}

		if err != nil {
			xcontext.Logger(ctx).Warnf("Cannot get receipt for tx hash %s: %v", tx.Hash().String(), err)
		}

		if ok {
			retry = 0
			txQueue = txQueue[1:]
		} else {
			if retry == MaxReceiptRetry {
				xcontext.Logger(ctx).Errorf("Cannot get receipt for tx with hash %s on chain %s",
					tx.Hash().String(), rf.chain)
				txQueue = txQueue[1:]
			} else {
				retry++
				time.Sleep(rf.retryTime)
			}
		}
	}

	return response
}

func (rf *defaultReceiptFetcher) fetchReceipts(ctx context.Context, block int64, txs []*types.TransactionWithOpts) {
	rf.requestCh <- &txReceiptRequest{
		blockNumber: block,
		txs:         txs,
	}
}
