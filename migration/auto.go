package migration

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

// When this migrator is called, no need to call other migrators.
func AutoMigrate(ctx context.Context) error {
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
		&entity.UserAggregate{},
		&entity.File{},
		&entity.Badge{},
		&entity.GameMap{},
		&entity.GameRoom{},
		&entity.GameUser{},
		&entity.Transaction{},
		&entity.Migration{},
	)
}
