package types

import "github.com/ethereum/go-ethereum/common"

type TrackResult int

const (
	TrackResultConfirmed TrackResult = iota
	TrackResultFailure
	TrackResultTimeout
)

type TrackUpdate struct {
	Chain       string
	Bytes       []byte
	BlockHeight int64
	Result      TrackResult
	Hash        common.Hash

	// For ETH
	Nonce int64
}
