package migration

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

var migrators = []func(context.Context) error{
	migrate0000,
	migrate0001,
	migrate0002,
	migrate0003,
	migrate0004,
	migrate0005,
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
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			// If not found any migration version, begin from version 1.
			currentVersion = 1
		} else {
			currentVersion = migration.Version + 1
		}
	}

	if currentVersion == 0 {
		// This migration version will create the database with the latest
		// version.
		if err := migrate0000(ctx); err != nil {
			return err
		}

		if len(migrators) > 1 {
			// Update the database version to the latest one.
			if err := db.Create(&entity.Migration{Version: len(migrators) - 1}).Error; err != nil {
				return err
			}
		}

		xcontext.Logger(ctx).Infof("Migrate all successfully")
		return nil
	}

	if currentVersion >= len(migrators) {
		xcontext.Logger(ctx).Infof("Database is up to date")
		return nil
	}

	xcontext.Logger(ctx).Infof("Begin migrating from version %d", currentVersion)
	for version := currentVersion; version < len(migrators); version++ {
		if err := migrators[version](ctx); err != nil {
			return err
		}

		if err := db.Create(&entity.Migration{Version: version}).Error; err != nil {
			return err
		}

		xcontext.Logger(ctx).Infof("Migrate version %d successfully", version)
	}
	xcontext.Logger(ctx).Infof("Migration completed")

	return nil
}
