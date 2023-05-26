package migration

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func migrate0002(ctx context.Context) error {
	migrator := xcontext.DB(ctx).Migrator()

	if err := migrator.AddColumn(&entity.Follower{}, "quests"); err != nil {
		return nil
	}

	if err := migrator.RenameColumn(&entity.Follower{}, "streak", "streaks"); err != nil {
		return nil
	}

	if err := migrator.AddColumn(&entity.Quest{}, "points"); err != nil {
		return nil
	}

	if err := migrator.DropTable("user_aggregates"); err != nil {
		return nil
	}

	return nil
}
