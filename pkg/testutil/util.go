package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gorilla/sessions"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/migration"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/token"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type redisClientKey struct{}

func NewQuestFactory(ctx context.Context) questclaim.Factory {
	return questclaim.NewFactory(
		repository.NewClaimedQuestRepository(),
		repository.NewQuestRepository(&MockSearchCaller{}),
		repository.NewCommunityRepository(&MockSearchCaller{}, RedisClient(ctx)),
		repository.NewFollowerRepository(),
		repository.NewOAuth2Repository(),
		repository.NewUserRepository(RedisClient(ctx)),
		repository.NewPayRewardRepository(),
		repository.NewBlockChainRepository(),
		repository.NewLotteryRepository(),
		repository.NewNftRepository(),
		&MockTwitterEndpoint{}, &MockDiscordEndpoint{},
		nil,
	)
}

func NewCommunityRoleVerifier(ctx context.Context) *common.CommunityRoleVerifier {
	return common.NewCommunityRoleVerifier(
		repository.NewFollowerRoleRepository(),
		repository.NewRoleRepository(),
		repository.NewUserRepository(RedisClient(ctx)),
	)
}

func MockContext(t *testing.T) context.Context {
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

	s := miniredis.RunT(t)
	client, err := xredis.NewClientWithCustomAddress(ctx, s.Addr())
	if err != nil {
		panic(err)
	}
	ctx = context.WithValue(ctx, redisClientKey{}, client)

	return ctx
}

func MockContextWithUserID(t *testing.T, userID string) context.Context {
	return xcontext.WithRequestUserID(MockContext(t), userID)
}

func RedisClient(ctx context.Context) xredis.Client {
	client := ctx.Value(redisClientKey{})
	if client == nil {
		return nil
	}

	return client.(xredis.Client)
}
