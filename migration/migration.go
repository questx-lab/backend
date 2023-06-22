package migration

import (
	"context"
	"embed"
	"io/fs"
	"os"
	"path/filepath"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

//go:embed mysql/*
var mysqlFS embed.FS

// MigrationsTempDir creates a temporary directory, populates it with the
// migration files, and returns the path to that directory.
// This is useful to run database migrations with only the binary without having
// to ship around the migration files separately.
//
// It is the caller's repsonsibility to remove the directory when it is no
// longer needed.
func MigrationsTempDir() (string, error) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}

	mFS, err := fs.Sub(mysqlFS, "mysql")
	if err != nil {
		return "", err
	}

	if err := fs.WalkDir(mFS, ".", func(path string, d fs.DirEntry, _ error) error {
		dst := filepath.Join(tmpDir, path)
		if dst == tmpDir {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		content, err := mysqlFS.ReadFile(filepath.Join("mysql", path))
		if err != nil {
			return err
		}

		return os.WriteFile(dst, content, 0600)
	}); err != nil {
		return "", err
	}

	return tmpDir, nil
}

func Migrate(ctx context.Context) error {
	if err := xcontext.DB(ctx).Migrator().CreateTable(&entity.GameLuckyboxEvent{}); err != nil {
		return err
	}
	if err := xcontext.DB(ctx).Migrator().CreateTable(&entity.GameLuckybox{}); err != nil {
		return err
	}

	return nil

	// db, err := xcontext.DB(ctx).DB()
	// if err != nil {
	// 	return err
	// }

	// migrationDir, err := MigrationsTempDir()
	// if err != nil {
	// 	return err
	// }

	// driver, err := mysql.WithInstance(db, &mysql.Config{})
	// if err != nil {
	// 	return err
	// }

	// m, err := migrate.NewWithDatabaseInstance(
	// 	"file://"+migrationDir, xcontext.Configs(ctx).Database.Database, driver)
	// if err != nil {
	// 	return err
	// }

	// if err := m.Up(); !errors.Is(err, migrate.ErrNoChange) {
	// 	return err
	// }

	// return nil
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
		&entity.BadgeDetail{},
		&entity.GameMap{},
		&entity.GameRoom{},
		&entity.GameUser{},
		&entity.Migration{},
		&entity.PayReward{},
	)
}
