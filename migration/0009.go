package migration

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func migrate0009(ctx context.Context) error {
	if err := xcontext.DB(ctx).Migrator().CreateTable(&entity.BlockChainTransaction{}); err != nil {
		return err
	}

	return nil
}
