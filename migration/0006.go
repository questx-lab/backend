package migration

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func migrate0006(ctx context.Context) error {
	return xcontext.DB(ctx).Migrator().CreateConstraint(&entity.Category{}, "CreatedByUser")
}
