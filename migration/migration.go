package migration

import (
	"context"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func Migrate(ctx context.Context) error {
	db, err := xcontext.DB(ctx).DB()
	if err != nil {
		return err
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return err
	}

	dbCfg := xcontext.Configs(ctx).Database
	m, err := migrate.NewWithDatabaseInstance("file://"+dbCfg.MigrationDir, dbCfg.Database, driver)
	if err != nil {
		return err
	}

	return m.Up()
}

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
		&entity.File{},
		&entity.Badge{},
		&entity.GameMap{},
		&entity.GameRoom{},
		&entity.GameUser{},
		&entity.Migration{},
		&entity.PayReward{},
	)
}
