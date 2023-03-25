package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

func DefaultTestDb() *gorm.DB {
	file := "/Users/billy/Desktop/code/crypto-projects/questx/backend/pkg/testutil/fixture/test.db"
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory", file)), &gorm.Config{})
	if err != nil {
		panic(err)
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
