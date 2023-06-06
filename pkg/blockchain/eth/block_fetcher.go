package eth

import (
	"context"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/ethereum/go-ethereum"
	etypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/pkg/math"
)

const (
	MinWaitTime = 500 // 500ms
)

type defaultBlockFetcher struct {
	blockHeight int64
	blockTime   int
	cfg         config.ChainConfig
	client      EthClient
	blockCh     chan *etypes.Block
}

func newBlockFetcher(cfg config.ChainConfig, blockCh chan *etypes.Block, client EthClient) *defaultBlockFetcher {
	return &defaultBlockFetcher{
		blockCh:   blockCh,
		cfg:       cfg,
		client:    client,
		blockTime: cfg.BlockTime,
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
			xcontext.Logger(ctx).Errorf("cannot get latest block number for chain %s. Sleeping for a few seconds\n", bf.cfg.Chain)
			time.Sleep(time.Second * 5)
			continue
		}

		bf.blockHeight = math.MaxInt64(int64(number)-int64(bf.cfg.ThresholdUpdateBlock), 0)
		break
	}

	xcontext.Logger(ctx).Infof("Watching from block ", bf.blockHeight, " for chain ", bf.cfg.Chain)
}

func (bf *defaultBlockFetcher) scanBlocks(ctx context.Context) {
	latestBlock, err := bf.getLatestBlock(ctx)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Failed to scan blocks, err = ", err)
	}

	if latestBlock != nil {
		bf.blockHeight = math.MaxInt64(latestBlock.Header().Number.Int64()-int64(bf.cfg.ThresholdUpdateBlock), 0)
	}
	xcontext.Logger(ctx).Infof(bf.cfg.Chain, " Latest height = ", bf.blockHeight)

	for {
		xcontext.Logger(ctx).Infof("Block time on chain ", bf.cfg.Chain, " is ", bf.blockTime)
		if bf.blockTime < 0 {
			bf.blockTime = 0
		}

		// Get the blockheight
		block, err := bf.tryGetBlock(ctx)
		if err != nil || block == nil {
			if _, ok := err.(*BlockHeightExceededError); !ok && err != ethereum.NotFound {
				// This err is not ETH not found or our custom error.
				xcontext.Logger(ctx).Errorf("Cannot get block at height %d for chain %s, err = %s\n",
					bf.blockHeight, bf.cfg.Chain, err)

				// Bug only on polygon network https://github.com/maticnetwork/bor/issues/387
				// The block exists but its header hash is equivalent to empty root hash but the internal
				// block has some transaction inside. Geth client throws an error in this situation.
				// This rarely happens but it does happen. Skip this block for now.
				if strings.Contains(bf.cfg.Chain, "polygon") &&
					strings.Contains(err.Error(), "server returned non-empty transaction list but block header indicates no transactions") {
					log.Printf("server returned non-empty transaction at block height %d in chain %s\n", bf.blockHeight, bf.cfg.Chain)
					bf.blockHeight = bf.blockHeight + 1
				}
			}

			bf.blockTime = bf.blockTime + bf.cfg.AdjustTime
			time.Sleep(time.Duration(bf.blockTime) * time.Millisecond)
			continue
		}

		bf.blockCh <- block
		bf.blockHeight++

		if bf.blockTime-bf.cfg.AdjustTime/4 > MinWaitTime {
			bf.blockTime = bf.blockTime - bf.cfg.AdjustTime/4
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
	ctx, cancel := context.WithTimeout(ctx, RpcTimeOut)
	defer cancel()

	return bf.client.BlockByNumber(ctx, blockNum)
}

// Get block with retry when block is not mined yet.
func (bf *defaultBlockFetcher) tryGetBlock(ctx context.Context) (*etypes.Block, error) {
	number, err := bf.getBlockNumber(ctx)
	if err != nil {
		return nil, err
	}

	if number-uint64(bf.cfg.ThresholdUpdateBlock) < uint64(bf.blockHeight) {
		return nil, NewBlockHeightExceededError(number)
	}

	block, err := bf.getBlock(ctx, bf.blockHeight)
	switch err {
	case nil:
		xcontext.Logger(ctx).Infof(bf.cfg.Chain, " Height = ", block.Number())
		if bf.blockHeight > 0 && number-uint64(bf.blockHeight) > 5 {
			bf.blockTime = MinWaitTime
		}
		return block, nil

	case ethereum.NotFound:
		// Sleep a few seconds and to get the block again.
		time.Sleep(time.Duration(math.MinInt(bf.blockTime/4, 3000)) * time.Millisecond)
		block, err = bf.getBlock(ctx, bf.blockHeight)

		// Extend the wait time a little bit more
		bf.blockTime = bf.blockTime + bf.cfg.AdjustTime
		xcontext.Logger(ctx).Infof("New blocktime: ", bf.blockTime)
	}

	return block, err
}

func (bf *defaultBlockFetcher) getBlockNumber(ctx context.Context) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, RpcTimeOut)
	defer cancel()

	return bf.client.BlockNumber(ctx)
}
