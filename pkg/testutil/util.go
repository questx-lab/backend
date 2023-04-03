package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/xcontext"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewMockContext() xcontext.Context {
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

	return xcontext.NewContext(context.Background(), r, nil, cfg, logger.NewLogger(), db)
}

func NewMockContextWithUserID(ctx xcontext.Context, userID string) xcontext.Context {
	if ctx == nil {
		ctx = NewMockContext()
	}

	xcontext.SetRequestUserID(ctx, userID)
	return ctx
}
