package migration

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func migrate0005(ctx context.Context) error {
	return xcontext.DB(ctx).Migrator().RenameColumn(&entity.ClaimedQuest{}, "input", "submission_data")
}
