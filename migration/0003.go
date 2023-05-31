package migration

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func migrate0003(ctx context.Context) error {
	if err := xcontext.DB(ctx).Migrator().AddColumn(&entity.Community{}, "display_name"); err != nil {
		return err
	}

	if err := xcontext.DB(ctx).Migrator().RenameColumn(&entity.Community{}, "name", "handle"); err != nil {
		return err
	}

	return nil
}
