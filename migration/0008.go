package migration

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func migrate0008(ctx context.Context) error {
	if err := xcontext.DB(ctx).Migrator().DropTable("transactions"); err != nil {
		return err
	}

	if err := xcontext.DB(ctx).Migrator().CreateTable(&entity.PayReward{}); err != nil {
		return err
	}
	return nil
}
