package migration

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func migrate0008(ctx context.Context) error {
	if err := xcontext.DB(ctx).Migrator().AddColumn(&entity.Transaction{}, "tx_hash"); err != nil {
		return err
	}

	return nil
}
