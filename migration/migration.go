package migration

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/migration/cql"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/scylladb/gocqlx/v2"
	scylladb_migrate "github.com/scylladb/gocqlx/v2/migrate"
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

func Migrate(ctx context.Context, twitterEndpoint twitter.IEndpoint) error {
	db, err := xcontext.DB(ctx).DB()
	if err != nil {
		return err
	}

	migrationDir, err := MigrationsTempDir()
	if err != nil {
		return err
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationDir, xcontext.Configs(ctx).Database.Database, driver)
	if err != nil {
		return err
	}

	oldVersion, dirty, err := m.Version()
	if err != nil {
		return err
	}

	if dirty {
		return fmt.Errorf("database is dirty at version %d", oldVersion)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	if err == nil { // If not ErrNoChange
		switch {
		case oldVersion < 27:
			xcontext.Logger(ctx).Infof("Begin back-compatible for migration 27")
			if err := BackCompatibleVersion27(ctx, twitterEndpoint); err != nil {
				return err
			}
		}
	}

	return nil
}

// BackCompatibleVersion27 indexes all categories.
func BackCompatibleVersion27(ctx context.Context, twitterEndpoint twitter.IEndpoint) error {
	var categoies []entity.Category
	if err := xcontext.DB(ctx).Find(&categoies).Error; err != nil {
		return err
	}

	positionMap := map[string]int{}
	for _, category := range categoies {
		position := positionMap[category.CommunityID.String]
		err := xcontext.DB(ctx).Model(&entity.Category{}).
			Where("id=?", category.ID).
			Update("position", position).Error
		if err != nil {
			return err
		}

		positionMap[category.CommunityID.String]++
	}

	return nil
}

func AutoMigrate(ctx context.Context) error {
	return xcontext.DB(ctx).AutoMigrate(
		&entity.User{},
		&entity.OAuth2{},
		&entity.Community{},
		&entity.Quest{},
		&entity.Category{},
		&entity.ClaimedQuest{},
		&entity.Follower{},
		&entity.FollowerRole{},
		&entity.APIKey{},
		&entity.RefreshToken{},
		&entity.File{},
		&entity.Badge{},
		&entity.BadgeDetail{},
		&entity.Migration{},
		&entity.PayReward{},
		&entity.Role{},
	)
}

func MigrateScyllaDB(ctx context.Context, session gocqlx.Session) error {
	// First run prints data
	if err := scylladb_migrate.FromFS(ctx, session, cql.Files); err != nil {
		return err
	}

	// Second run skips the processed files
	if err := scylladb_migrate.FromFS(ctx, session, cql.Files); err != nil {
		return err
	}

	return nil
}
