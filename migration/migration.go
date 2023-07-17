package migration

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/api/twitter"
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

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	if err == nil { // If not ErrNoChange
		version, dirty, err := m.Version()
		if dirty {
			return errors.New("database is dirty")
		}

		if err != nil {
			return err
		}

		switch version {
		case 15:
			xcontext.Logger(ctx).Infof("Begin back-compatible for migration 15")
			if err := BackCompatibleVersion15(ctx, twitterEndpoint); err != nil {
				return err
			}
			fallthrough
		case 16:
			xcontext.Logger(ctx).Infof("Begin back-compatible for migration 16")
			if err := BackCompatibleVersion16(ctx, twitterEndpoint); err != nil {
				return err
			}
		}
	}

	return nil
}

// BackCompatibleVersion15 converts id of twitter oauth2 records to username instead.
// Before this version, we was using username of twitter as id.
func BackCompatibleVersion15(ctx context.Context, twitterEndpoint twitter.IEndpoint) error {
	var oauth2Users []entity.OAuth2
	if err := xcontext.DB(ctx).Find(&oauth2Users, "service=?", "twitter").Error; err != nil {
		return err
	}

	for _, oauth2User := range oauth2Users {
		if oauth2User.ServiceUsername != "" {
			xcontext.Logger(ctx).Debugf("Ignore user %s", oauth2User.UserID)
			continue
		}

		tag, username, found := strings.Cut(oauth2User.ServiceUserID, "_")
		if !found || tag != xcontext.Configs(ctx).Auth.Twitter.Name {
			return fmt.Errorf("unknown twitter tag of user %s", oauth2User.UserID)
		}

		user, err := twitterEndpoint.GetUser(ctx, username)
		if err != nil {
			return err
		}

		err = xcontext.DB(ctx).Model(&entity.OAuth2{}).
			Where("user_id=? AND service=?", oauth2User.UserID, "twitter").
			Updates(map[string]any{
				"service_user_id":  fmt.Sprintf("twitter_%s", user.ID),
				"service_username": user.ScreenName,
			}).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// BackCompatibleVersion16 indexes all quests.
func BackCompatibleVersion16(ctx context.Context, twitterEndpoint twitter.IEndpoint) error {
	var quests []entity.Quest
	if err := xcontext.DB(ctx).Find(&quests).Error; err != nil {
		return err
	}

	positionMap := map[string]int{}
	for _, quest := range quests {
		position := positionMap[quest.CategoryID.String]
		err := xcontext.DB(ctx).Model(&entity.Quest{}).
			Where("id=?", quest.ID).
			Update("position", position).Error
		if err != nil {
			return err
		}

		positionMap[quest.CategoryID.String]++
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
		&entity.Role{},
	)
}
