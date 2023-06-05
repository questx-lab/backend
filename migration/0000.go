package migration

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

// migrate0000 will create the database with the latest version.
func migrate0000(ctx context.Context) error {
	return xcontext.DB(ctx).AutoMigrate(
		&entity.User{},
		&entity.OAuth2{},
		&entity.Community{},
		&entity.Quest{},
		&entity.Collaborator{},
		&entity.Category{},
		&entity.ClaimedQuest{},
		&entity.Follower{},
		&entity.APIKey{},
		&entity.RefreshToken{},
		&entity.File{},
		&entity.Badge{},
		&entity.GameMap{},
		&entity.GameRoom{},
		&entity.GameUser{},
		&entity.Migration{},
	)
}
