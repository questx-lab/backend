package testutil

import (
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
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/router"
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
	if err != nil {
		t.Fail()
		return nil
	}

	data := string(bz)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
		return nil
	}

	tx := db.Exec(data)
	if tx.Error != nil {
		t.Fatal(tx.Error)
		return nil
	}

	return db
}

func NewMockContextWithUserID(userID string) router.Context {
	ctx := router.DefaultContext()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	tokenEngine := authenticator.NewTokenEngine[model.AccessToken](config.TokenConfigs{
		Secret:     "secret",
		Expiration: time.Minute,
	})

	tkn, _ := tokenEngine.Generate(userID, model.AccessToken{
		ID: userID,
	})
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tkn))
	ctx.SetRequest(r)
	ctx.SetAccessTokenEngine(tokenEngine)
	return ctx
}
