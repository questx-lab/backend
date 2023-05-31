package mocks

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type EthClient struct {
	mock.Mock
}

func (c *EthClient) Start() {
}

func (c *EthClient) BlockNumber(arg1 context.Context) (uint64, error) {
	args := c.Called(arg1)

	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	return args.Get(0).(uint64), args.Error(1)
}

func (c *EthClient) BlockByNumber(arg1 context.Context, arg2 *big.Int) (*ethtypes.Block, error) {
	args := c.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ethtypes.Block), args.Error(1)
}

func (c *EthClient) TransactionReceipt(arg1 context.Context, arg2 common.Hash) (*ethtypes.Receipt, error) {
	args := c.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ethtypes.Receipt), args.Error(1)
}

func (c *EthClient) SuggestGasPrice(arg1 context.Context) (*big.Int, error) {
	args := c.Called(arg1)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (c *EthClient) PendingNonceAt(arg1 context.Context, arg2 common.Address) (uint64, error) {
	args := c.Called(arg1, arg2)

	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	return args.Get(0).(uint64), args.Error(1)
}

func (c *EthClient) SendTransaction(arg1 context.Context, arg2 *ethtypes.Transaction) error {
	args := c.Called(arg1)

	return args.Error(0)
}

func (c *EthClient) BalanceAt(arg1 context.Context, arg2 common.Address, arg3 *big.Int) (*big.Int, error) {
	args := c.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}
