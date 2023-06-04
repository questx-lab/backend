package migration

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func migrate0007(ctx context.Context) error {
	if err := xcontext.DB(ctx).Migrator().DropColumn(&entity.Community{}, "shared_content_types"); err != nil {
		return err
	}

	if err := xcontext.DB(ctx).Migrator().DropColumn(&entity.Community{}, "team_size"); err != nil {
		return err
	}

	if err := xcontext.DB(ctx).Migrator().DropColumn(&entity.Community{}, "development_stage"); err != nil {
		return err
	}

	return nil
}
