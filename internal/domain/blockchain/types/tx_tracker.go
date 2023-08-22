package types

import (
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type TrackResult int

const (
	TrackResultConfirmed TrackResult = iota
	TrackResultFailure
	TrackResultTimeout
)

type TransactionWithOpts struct {
	*ethtypes.Transaction
	Opts string
}

type TrackUpdate struct {
	Chain       string
	Bytes       []byte
	BlockHeight int64
	Result      TrackResult
	Hash        common.Hash
	Opts        string

	// For ETH
	Nonce int64
}
