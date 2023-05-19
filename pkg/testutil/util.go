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
	return NewMockContextWith(httptest.NewRequest(http.MethodGet, "/", nil))
}

func NewMockContextWith(r *http.Request) xcontext.Context {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err := entity.MigrateTable(db); err != nil {
		panic(err)
	}

	cfg := config.Configs{
		ApiServer: config.APIServerConfigs{
			MaxLimit:     50,
			DefaultLimit: 1,
		},
		Auth: config.AuthConfigs{
			TokenSecret: "secret",
			AccessToken: config.TokenConfigs{
				Name:       "access_token",
				Expiration: time.Minute,
			},
			RefreshToken: config.TokenConfigs{
				Name:       "refresh_token",
				Expiration: time.Minute,
			},
		},
		Quest: config.QuestConfigs{
			QuizMaxQuestions:               10,
			QuizMaxOptions:                 10,
			InviteProjectRequiredFollowers: 1,
			InviteProjectRewardToken:       "USDT",
			InviteProjectRewardAmount:      10,
		},
	}

	return xcontext.NewContext(context.Background(), r, nil, cfg, logger.NewLogger(), db, nil)
}

func NewMockContextWithUserID(ctx xcontext.Context, userID string) xcontext.Context {
	if ctx == nil {
		ctx = NewMockContext()
	}

	xcontext.SetRequestUserID(ctx, userID)
	return ctx
}
