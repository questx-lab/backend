package testutil

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/questx-lab/backend/internal/repository/migration"
)

func GetEmptyTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("Failed to create in memory db")
	}

	err = migration.DoSqlMigration(db)
	if err != nil {
		t.Fatal("Failed to migrate db")
	}

	return db
}

func GetEmptyIntegrationDb(t *testing.T) *sql.DB {
	db, err := sql.Open("mysql", os.Getenv("DB_CONNECTION"))
	if err != nil {
		t.Fatal("cannot connect to db")
	}

	// TODO: Drop all the schema here & do reset db here.

	err = migration.DoSqlMigration(db)
	if err != nil {
		t.Fatal("cannot do migration for integration test")
	}

	return db
}

func EnableIntegrationTest() bool {
	return len(os.Getenv("RUN_INTEGRATION_TEST")) > 0
}
