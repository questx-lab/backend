package repository

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migration/*
var migrationsFS embed.FS

// MigrationsTempDir creates a temporary directory, populates it with the migration files,
// and returns the path to that directory.
// This is useful to run database migrations with only the dheart binary,
// without having to ship around the migration files separately.
//
// It is the caller's repsonsibility to remove the directory when it is no longer needed.
func MigrationsTempDir() (string, error) {
	tmpDir, err := os.MkdirTemp("", "migrations-*")
	if err != nil {
		return "", err
	}

	mFS, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return "", err
	}

	if err := fs.WalkDir(mFS, ".", func(path string, d fs.DirEntry, _ error) error {
		dst := filepath.Join(tmpDir, path)
		if dst == tmpDir {
			return nil
		}

		if d.IsDir() {
			if err := os.Mkdir(dst, 0700); err != nil {
				return fmt.Errorf("failed to mkdir %q: %w", dst, err)
			}
			return nil
		}

		content, err := migrationsFS.ReadFile(filepath.Join("migrations", path))
		if err != nil {
			return err
		}

		return os.WriteFile(dst, content, 0600)
	}); err != nil {
		return "", err
	}

	return tmpDir, nil
}

type dbLogger struct {
}

func (loggger *dbLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func (loggger *dbLogger) Verbose() bool {
	return true
}

// doSqlMigration does sql migration in external database using "golang-migrate/migrate" lib.
func DoSqlMigration(db *sql.DB) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return err
	}

	// Write the migrations to a temporary directory
	// so they don't need to be managed out of band from the dheart binary.
	migrationDir, err := MigrationsTempDir()
	if err != nil {
		return fmt.Errorf("failed to create temporary directory for migrations: %w", err)
	}
	defer os.RemoveAll(migrationDir)

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationDir,
		"mysql",
		driver,
	)

	if err != nil {
		return err
	}

	m.Log = &dbLogger{}
	m.Up()

	return nil
}
