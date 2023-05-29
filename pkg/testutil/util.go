package testutil

import (
	"context"
	"time"

	"github.com/gorilla/sessions"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/migration"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/xcontext"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func MockContext() context.Context {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
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
		Session: config.SessionConfigs{
			Secret: "session-secret",
		},
		Quest: config.QuestConfigs{
			QuizMaxQuestions:                 10,
			QuizMaxOptions:                   10,
			InviteCommunityRequiredFollowers: 1,
			InviteCommunityRewardToken:       "USDT",
			InviteCommunityRewardAmount:      10,
		},
	}

	ctx := context.Background()
	ctx = xcontext.WithConfigs(ctx, cfg)
	ctx = xcontext.WithLogger(ctx, logger.NewLogger())
	ctx = xcontext.WithTokenEngine(ctx, authenticator.NewTokenEngine(cfg.Auth.TokenSecret))
	ctx = xcontext.WithSessionStore(ctx, sessions.NewCookieStore([]byte(cfg.Session.Secret)))
	ctx = xcontext.WithDB(ctx, db)

	if err := migration.Migrate(ctx); err != nil {
		panic(err)
	}

	return ctx
}

func MockContextWithUserID(userID string) context.Context {
	return xcontext.WithRequestUserID(MockContext(), userID)
}
