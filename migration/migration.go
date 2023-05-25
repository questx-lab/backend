package migration

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

var Migrators = []func(context.Context) error{
	AutoMigrate,
}

func Migrate(ctx context.Context) error {
	db := xcontext.DB(ctx)
	var currentVersion int
	if !db.Migrator().HasTable(&entity.Migration{}) {
		currentVersion = 0
	} else {
		// Find the last migration version, migrate next versions.
		migration := entity.Migration{}
		if err := db.Last(&migration).Error; err != nil {
			return err
		}

		currentVersion = migration.Version + 1
	}

	xcontext.Logger(ctx).Infof("Begin migrating from version %d", currentVersion)
	hasCalledAutoMigration := false
	for version := currentVersion; version < len(Migrators); version++ {
		// If the first version of migration (auto migration) is called, no need
		// to call other versions.
		if !hasCalledAutoMigration {
			if err := Migrators[version](ctx); err != nil {
				return err
			}

			if version == 0 {
				hasCalledAutoMigration = true
			}
		}

		if version == 0 {
			xcontext.Logger(ctx).Infof("Auto migration successfully")
		} else {
			// We still add all versions into database even if we didn't call
			// migrator of those versions.
			if err := db.Create(&entity.Migration{Version: version}).Error; err != nil {
				return err
			}

			if currentVersion > 0 {
				xcontext.Logger(ctx).Infof("Migrate version %d successfully", version)
			}
		}
	}
	xcontext.Logger(ctx).Infof("Migration completed")

	return nil
}
