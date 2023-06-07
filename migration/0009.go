package migration

import (
	"context"

	"github.com/questx-lab/backend/pkg/xcontext"
)

type BlockChainTransaction struct {
	Chain       string
	TxHash      string
	BlockHeight int64
	TxBytes     []byte
}

func migrate0009(ctx context.Context) error {
	if err := xcontext.DB(ctx).Migrator().CreateTable(&BlockChainTransaction{}); err != nil {
		return err
	}

	return nil
}
