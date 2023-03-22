package testutil

import (
	"testing"

	"github.com/questx-lab/backend/internal/repository/migration"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"gorm.io/driver/sqlite"
)

func GetTestDb(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.Nil(t, err)

	migration.DoMigration(db)

	return db
}
