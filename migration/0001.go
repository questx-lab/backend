package migration

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func Migrate0001(ctx context.Context) error {
	if err := xcontext.DB(ctx).Migrator().AlterColumn(&entity.Category{}, "community_id"); err != nil {
		return err
	}

	if err := xcontext.DB(ctx).Migrator().AlterColumn(&entity.Category{}, "name"); err != nil {
		return err
	}

	return nil
}
