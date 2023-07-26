package client

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/internal/domain/blockchain/types"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type BlockchainCaller interface {
	GetTokenInfo(ctx context.Context, chain, address string) (types.TokenInfo, error)
	Close()
}

type blockchainCaller struct {
	client *rpc.Client
}

func NewBlockchainCaller(client *rpc.Client) *blockchainCaller {
	return &blockchainCaller{client: client}
}

func (c *blockchainCaller) GetTokenInfo(ctx context.Context, chain, address string) (types.TokenInfo, error) {
	var result types.TokenInfo
	err := c.client.CallContext(ctx, &result, c.fname(ctx, "getTokenInfo"), chain, address)
	if err != nil {
		return types.TokenInfo{}, err
	}

	return result, nil
}

func (c *blockchainCaller) Close() {
	c.client.Close()
}

func (c *blockchainCaller) fname(ctx context.Context, funcName string) string {
	return fmt.Sprintf("%s_%s", xcontext.Configs(ctx).Blockchain.RPCName, funcName)
}
