package testutil

import (
	"context"
	"time"

	"github.com/gorilla/sessions"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/migration"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/token"
	"github.com/questx-lab/backend/pkg/xcontext"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var QuestFactory = questclaim.NewFactory(
	repository.NewClaimedQuestRepository(),
	repository.NewQuestRepository(&MockSearchCaller{}),
	repository.NewCommunityRepository(&MockSearchCaller{}),
	repository.NewFollowerRepository(),
	repository.NewOAuth2Repository(),
	repository.NewUserRepository(&MockRedisClient{}),
	repository.NewPayRewardRepository(),
	repository.NewBlockChainRepository(),
	repository.NewLotteryRepository(),
	&MockTwitterEndpoint{}, &MockDiscordEndpoint{},
	nil,
)

var CommunityRoleVerifier = common.NewCommunityRoleVerifier(
	repository.NewFollowerRoleRepository(),
	repository.NewRoleRepository(),
	repository.NewUserRepository(&MockRedisClient{}),
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
			QuizMaxQuestions:                  10,
			QuizMaxOptions:                    10,
			InviteCommunityRequiredFollowers:  1,
			InviteCommunityRewardChain:        "avaxc-testnet",
			InviteCommunityRewardTokenAddress: "USDT",
			InviteCommunityRewardAmount:       10,
		},
	}

	ctx := context.Background()
	ctx = xcontext.WithConfigs(ctx, cfg)
	ctx = xcontext.WithLogger(ctx, logger.NewLogger(logger.DEBUG))
	ctx = xcontext.WithTokenEngine(ctx, token.NewEngine(cfg.Auth.TokenSecret))
	ctx = xcontext.WithSessionStore(ctx, sessions.NewCookieStore([]byte(cfg.Session.Secret)))
	ctx = xcontext.WithDB(ctx, db)

	if err := migration.AutoMigrate(ctx); err != nil {
		panic(err)
	}

	return ctx
}

func MockContextWithUserID(userID string) context.Context {
	return xcontext.WithRequestUserID(MockContext(), userID)
}
