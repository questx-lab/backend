package migration

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

var Migrators = []func(context.Context) error{}

func Migrate(ctx context.Context) error {
	db := xcontext.DB(ctx)
	var currentVersion int
	if !db.Migrator().HasTable(&entity.Migration{}) {
		currentVersion = -1
	} else {
		// Find the last migration version, migrate next versions.
		migration := entity.Migration{}
		if err := db.Last(&migration).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			// If not found any migration version, begin from version 1.
			currentVersion = 0
		} else {
			currentVersion = migration.Version
		}
	}

	if currentVersion == -1 {
		if err := AutoMigrate(ctx); err != nil {
			return err
		}

		if len(Migrators) > 0 {
			if err := db.Create(&entity.Migration{Version: len(Migrators)}).Error; err != nil {
				return err
			}
		}

		xcontext.Logger(ctx).Infof("Auto migrate successfully")
		return nil
	}

	xcontext.Logger(ctx).Infof("Begin migrating from version %d", currentVersion)
	for version := currentVersion; version < len(Migrators); version++ {
		if err := Migrators[version](ctx); err != nil {
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
