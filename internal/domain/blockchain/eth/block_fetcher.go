package eth

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/ethereum/go-ethereum"
	etypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/pkg/math"
)

const (
	MinWaitTime = 500 // 500ms
)

type defaultBlockFetcher struct {
	chain                string
	blockHeight          int64
	adjustTime           int
	blockTime            int
	thresholdUpdateBlock int
	client               EthClient
	blockCh              chan *etypes.Block
}

func newBlockFetcher(blockchain *entity.Blockchain, blockCh chan *etypes.Block, client EthClient) *defaultBlockFetcher {
	return &defaultBlockFetcher{
		chain:                blockchain.Name,
		blockCh:              blockCh,
		client:               client,
		blockTime:            blockchain.BlockTime,
		adjustTime:           blockchain.AdjustTime,
		thresholdUpdateBlock: blockchain.ThresholdUpdateBlock,
	}
}

func (bf *defaultBlockFetcher) start(ctx context.Context) {
	bf.setBlockHeight(ctx)
	bf.scanBlocks(ctx)
}

func (bf *defaultBlockFetcher) setBlockHeight(ctx context.Context) {
	for {
		number, err := bf.getBlockNumber(ctx)
		if err != nil {
			xcontext.Logger(ctx).Errorf(
				"Cannot get latest block number for chain %s. Sleeping for a few seconds", bf.chain)
			time.Sleep(time.Second * 5)
			continue
		}

		bf.blockHeight = math.MaxInt64(int64(number)-int64(bf.thresholdUpdateBlock), 0)
		break
	}

	xcontext.Logger(ctx).Infof("Watching from block %d for chain %s", bf.blockHeight, bf.chain)
}

func (bf *defaultBlockFetcher) scanBlocks(ctx context.Context) {
	latestBlock, err := bf.getLatestBlock(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Failed to scan blocks: %v", err)
	}

	if latestBlock != nil {
		bf.blockHeight = math.MaxInt64(latestBlock.Header().Number.Int64()-int64(bf.thresholdUpdateBlock), 0)
	}
	xcontext.Logger(ctx).Infof("%s Latest height = %d", bf.chain, bf.blockHeight)

	for {
		xcontext.Logger(ctx).Debugf("Block time on chain %s is %d", bf.chain, bf.blockTime)
		if bf.blockTime < 0 {
			bf.blockTime = 0
		}

		// Get the blockheight
		block, err := bf.tryGetBlock(ctx)
		if err != nil || block == nil {
			if _, ok := err.(*BlockHeightExceededError); !ok && err != ethereum.NotFound {
				// This err is not ETH not found or our custom error.
				xcontext.Logger(ctx).Errorf("Cannot get block at height %d for chain %s, err = %s",
					bf.blockHeight, bf.chain, err)

				// Bug only on polygon network https://github.com/maticnetwork/bor/issues/387
				// The block exists but its header hash is equivalent to empty root hash but the internal
				// block has some transaction inside. Geth client throws an error in this situation.
				// This rarely happens but it does happen. Skip this block for now.
				if strings.Contains(bf.chain, "polygon") &&
					strings.Contains(err.Error(), "server returned non-empty transaction list but block header indicates no transactions") {
					xcontext.Logger(ctx).Errorf(
						"Server returned non-empty transaction at block height %d in chain %s",
						bf.blockHeight, bf.chain)
					bf.blockHeight = bf.blockHeight + 1
				}
			}

			bf.blockTime = bf.blockTime + bf.adjustTime
			time.Sleep(time.Duration(bf.blockTime) * time.Millisecond)
			continue
		}

		bf.blockCh <- block
		bf.blockHeight++

		if bf.blockTime-bf.adjustTime/4 > MinWaitTime {
			bf.blockTime = bf.blockTime - bf.adjustTime/4
		}
		time.Sleep(time.Duration(bf.blockTime) * time.Millisecond)
	}
}

func (bf *defaultBlockFetcher) getLatestBlock(ctx context.Context) (*etypes.Block, error) {
	return bf.getBlock(ctx, -1)
}

func (bf *defaultBlockFetcher) getBlock(ctx context.Context, height int64) (*etypes.Block, error) {
	blockNum := big.NewInt(height)
	if height == -1 { // latest block
		blockNum = nil
	}

	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, RpcTimeOut)
	defer cancel()

	return bf.client.BlockByNumber(ctx, blockNum)
}

// Get block with retry when block is not mined yet.
func (bf *defaultBlockFetcher) tryGetBlock(ctx context.Context) (*etypes.Block, error) {
	number, err := bf.getBlockNumber(ctx)
	if err != nil {
		return nil, err
	}

	if number-uint64(bf.thresholdUpdateBlock) < uint64(bf.blockHeight) {
		return nil, NewBlockHeightExceededError(number)
	}

	block, err := bf.getBlock(ctx, bf.blockHeight)
	switch err {
	case nil:
		xcontext.Logger(ctx).Infof("%s Height = %d", bf.chain, block.Number())
		if bf.blockHeight > 0 && number-uint64(bf.blockHeight) > 5 {
			bf.blockTime = MinWaitTime
		}
		return block, nil

	case ethereum.NotFound:
		// Sleep a few seconds and to get the block again.
		time.Sleep(time.Duration(math.MinInt(bf.blockTime/4, 3000)) * time.Millisecond)
		block, err = bf.getBlock(ctx, bf.blockHeight)

		// Extend the wait time a little bit more
		bf.blockTime = bf.blockTime + bf.adjustTime
		xcontext.Logger(ctx).Infof("New blocktime: %v", bf.blockTime)
	}

	return block, err
}

func (bf *defaultBlockFetcher) getBlockNumber(ctx context.Context) (uint64, error) {
	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, RpcTimeOut)
	defer cancel()

	return bf.client.BlockNumber(ctx)
}
