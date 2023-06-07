package migration

import (
	"context"

	"github.com/questx-lab/backend/pkg/xcontext"
)

// This struct is cloned from the first version of enity.PayReward. It is used
// to keep the original table structure. If we call CreateTable with
// entity.PayReward instead of this struct, it always creates a table with the
// latest version, and modifying columns of this table in the future may be
// failed.
type BlockChainTransaction struct {
	// NOTE: Please make sure this is the latest version Base at the time this
	// file is created.
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
