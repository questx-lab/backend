package testutil

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/router"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func GetEmptyTestDb() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if err := entity.MigrateTable(db); err != nil {
		panic(err)
	}

	return db
}

func getTestDbName() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	return filepath.Join(dir, DbDump)
}

func DefaultTestDb(t *testing.T) *gorm.DB {
	file := getTestDbName()
	bz, err := os.ReadFile(file)
	require.NoError(t, err)

	data := string(bz)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	tx := db.Exec(data)
	require.NoError(t, tx.Error)

	return db
}

func NewMockContext() router.Context {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err := entity.MigrateTable(db); err != nil {
		panic(err)
	}

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	cfg := config.Configs{
		Token: config.TokenConfigs{
			Secret:     "secret",
			Expiration: time.Minute,
		},
	}

	return router.NewContext(context.Background(), r, nil, cfg, logger.NewLogger(), db)
}

func NewMockContextWithUserID(ctx router.Context, userID string) router.Context {
	if ctx == nil {
		ctx = NewMockContext()
	}

	tkn, _ := ctx.AccessTokenEngine().Generate(userID, model.AccessToken{ID: userID})
	ctx.Request().Header.Add("Authorization", fmt.Sprintf("Bearer %s", tkn))
	return ctx
}
