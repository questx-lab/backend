package ethutil

import (
	cid "github.com/ipfs/go-cid"
	mc "github.com/multiformats/go-multicodec"
	mh "github.com/multiformats/go-multihash"
)

// Create a cid manually by specifying the 'prefix' parameters
var pref = cid.Prefix{
	Version:  1,
	Codec:    uint64(mc.Raw),
	MhType:   mh.SHA2_256,
	MhLength: -1, // default length
}

func GetIpfsHash(data []byte) ([]byte, error) {
	// And then feed it some data
	c, err := pref.Sum(data)
	if err != nil {
		return nil, err
	}

	return c.Bytes(), nil
}
