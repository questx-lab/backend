package client

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/questx-lab/backend/internal/domain/blockchain/types"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type BlockchainCaller interface {
	MintNFT(ctx context.Context, communityID, chain string, nftID int64, amount int) (string, error)
	ERC20TokenInfo(ctx context.Context, chain, address string) (types.TokenInfo, error)
	ERC20BalanceOf(ctx context.Context, chain, tokenAddress, accountAddress string) (*big.Int, error)
	ERC1155BalanceOf(ctx context.Context, chain, address string, tokenID int64) (*big.Int, error)
	DeployNFT(ctx context.Context, chain string) (string, error)
	Close()
}

type blockchainCaller struct {
	client *rpc.Client
}

func NewBlockchainCaller(client *rpc.Client) *blockchainCaller {
	return &blockchainCaller{client: client}
}

func (c *blockchainCaller) ERC20TokenInfo(ctx context.Context, chain, address string) (types.TokenInfo, error) {
	var result types.TokenInfo
	err := c.client.CallContext(ctx, &result, c.fname(ctx, "eRC20TokenInfo"), chain, address)
	if err != nil {
		return types.TokenInfo{}, err
	}

	return result, nil
}

func (c *blockchainCaller) ERC20BalanceOf(
	ctx context.Context, chain, tokenAddress, accountAddress string,
) (*big.Int, error) {
	var result *big.Int
	err := c.client.CallContext(ctx, &result, c.fname(ctx, "eRC20BalanceOf"), chain, tokenAddress, accountAddress)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *blockchainCaller) ERC1155BalanceOf(
	ctx context.Context, chain, address string, tokenID int64,
) (*big.Int, error) {
	var result *big.Int
	err := c.client.CallContext(ctx, &result, c.fname(ctx, "eRC1155BalanceOf"), chain, address, tokenID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *blockchainCaller) MintNFT(
	ctx context.Context, communityID, chain string, nftID int64, amount int,
) (string, error) {
	var result string
	err := c.client.CallContext(ctx, &result, c.fname(ctx, "mintNFT"), communityID, chain, nftID, amount)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (c *blockchainCaller) DeployNFT(ctx context.Context, chain string) (string, error) {
	var result string
	err := c.client.CallContext(ctx, &result, c.fname(ctx, "deployNFT"), chain)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (c *blockchainCaller) Close() {
	c.client.Close()
}

func (c *blockchainCaller) fname(ctx context.Context, funcName string) string {
	return fmt.Sprintf("%s_%s", xcontext.Configs(ctx).Blockchain.RPCName, funcName)
}
