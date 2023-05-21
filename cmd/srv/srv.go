package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/domain"
	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/domain/gameproxy"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/api/telegram"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/kafka"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type App struct {
	ctx context.Context

	userRepo          repository.UserRepository
	oauth2Repo        repository.OAuth2Repository
	projectRepo       repository.ProjectRepository
	questRepo         repository.QuestRepository
	categoryRepo      repository.CategoryRepository
	collaboratorRepo  repository.CollaboratorRepository
	claimedQuestRepo  repository.ClaimedQuestRepository
	participantRepo   repository.ParticipantRepository
	fileRepo          repository.FileRepository
	apiKeyRepo        repository.APIKeyRepository
	refreshTokenRepo  repository.RefreshTokenRepository
	userAggregateRepo repository.UserAggregateRepository
	gameRepo          repository.GameRepository
	badgeRepo         repository.BadgeRepo
	transactionRepo   repository.TransactionRepository

	userDomain         domain.UserDomain
	authDomain         domain.AuthDomain
	projectDomain      domain.ProjectDomain
	questDomain        domain.QuestDomain
	categoryDomain     domain.CategoryDomain
	collaboratorDomain domain.CollaboratorDomain
	claimedQuestDomain domain.ClaimedQuestDomain
	fileDomain         domain.FileDomain
	apiKeyDomain       domain.APIKeyDomain
	gameProxyDomain    domain.GameProxyDomain
	gameDomain         domain.GameDomain
	statisticDomain    domain.StatisticDomain
	participantDomain  domain.ParticipantDomain
	transactionDomain  domain.TransactionDomain

	publisher   pubsub.Publisher
	proxyRouter gameproxy.Router

	server *http.Server
	router *router.Router

	storage storage.Storage

	badgeManager     *badge.Manager
	twitterEndpoint  twitter.IEndpoint
	discordEndpoint  discord.IEndpoint
	telegramEndpoint telegram.IEndpoint
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		value = fallback
	}
	return value
}

func (app *App) loadConfig() config.Configs {
	return config.Configs{
		Env: getEnv("ENV", "local"),
		ApiServer: config.APIServerConfigs{
			MaxLimit:     parseInt(getEnv("API_MAX_LIMIT", "50")),
			DefaultLimit: parseInt(getEnv("API_DEFAULT_LIMIT", "1")),
			ServerConfigs: config.ServerConfigs{
				Host:      getEnv("API_HOST", "localhost"),
				Port:      getEnv("API_PORT", "8080"),
				AllowCORS: strings.Split(getEnv("API_ALLOW_CORS", "http://localhost:3000"), ","),
			},
		},
		GameProxyServer: config.ServerConfigs{
			Host:      getEnv("GAME_PROXY_HOST", "localhost"),
			Port:      getEnv("GAME_PROXY_PORT", "8081"),
			AllowCORS: strings.Split(getEnv("GAME_PROXY_ALLOW_CORS", "http://localhost:3000"), ","),
		},
		Auth: config.AuthConfigs{
			TokenSecret: getEnv("TOKEN_SECRET", "token_secret"),
			AccessToken: config.TokenConfigs{
				Name:       "access_token",
				Expiration: parseDuration(getEnv("ACCESS_TOKEN_DURATION", "5m")),
			},
			RefreshToken: config.TokenConfigs{
				Name:       "refresh_token",
				Expiration: parseDuration(getEnv("REFRESH_TOKEN_DURATION", "20m")),
			},
			Google: config.OAuth2Config{
				Name:      "google",
				VerifyURL: "https://www.googleapis.com/oauth2/v1/userinfo",
				IDField:   "email",
			},
			Twitter: config.OAuth2Config{
				Name:      "twitter",
				VerifyURL: "https://api.twitter.com/2/users/me",
				IDField:   "data.username",
			},
			Discord: config.OAuth2Config{
				Name:      "discord",
				VerifyURL: "https://discord.com/api/users/@me",
				IDField:   "id",
			},
			Telegram: config.TelegramConfigs{
				Name:            "telegram",
				BotToken:        getEnv("TELEGRAM_BOT_TOKEN", "telegram-bot-token"),
				LoginExpiration: parseDuration(getEnv("TELEGRAM_LOGIN_EXPIRATION", "10s")),
			},
		},
		Database: config.DatabaseConfigs{
			Host:     getEnv("MYSQL_HOST", "mysql"),
			Port:     getEnv("MYSQL_PORT", "3306"),
			User:     getEnv("MYSQL_USER", "mysql"),
			Password: getEnv("MYSQL_PASSWORD", "mysql"),
			Database: getEnv("MYSQL_DATABASE", "questx"),
			LogLevel: getEnv("DATABASE_LOG_LEVEL", "error"),
		},
		Session: config.SessionConfigs{
			Secret: getEnv("AUTH_SESSION_SECRET", "secret"),
			Name:   "auth_session",
		},
		Storage: storage.S3Configs{
			Region:    getEnv("STORAGE_REGION", "auto"),
			Endpoint:  getEnv("STORAGE_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("STORAGE_ACCESS_KEY", "access_key"),
			SecretKey: getEnv("STORAGE_SECRET_KEY", "secret_key"),
			Env:       getEnv("ENV", "local"),
		},
		File: config.FileConfigs{
			MaxSize: int64(parseEnvAsInt("MAX_UPLOAD_FILE", 2*1024*1024)),
		},
		Quest: config.QuestConfigs{
			Twitter: config.TwitterConfigs{
				ReclaimDelay:      parseDuration(getEnv("TWITTER_RECLAIM_DELAY", "15m")),
				AppAccessToken:    getEnv("TWITTER_APP_ACCESS_TOKEN", "app_access_token"),
				ConsumerAPIKey:    getEnv("TWITTER_CONSUMER_API_KEY", "consumer_key"),
				ConsumerAPISecret: getEnv("TWITTER_CONSUMER_API_SECRET", "comsumer_secret"),
				AccessToken:       getEnv("TWITTER_ACCESS_TOKEN", "access_token"),
				AccessTokenSecret: getEnv("TWITTER_ACCESS_TOKEN_SECRET", "access_token_secret"),
			},
			Dicord: config.DiscordConfigs{
				ReclaimDelay: parseDuration(getEnv("DISCORD_RECLAIM_DELAY", "15m")),
				BotToken:     getEnv("DISCORD_BOT_TOKEN", "discord_bot_token"),
				BotID:        getEnv("DISCORD_BOT_ID", "discord_bot_id"),
			},
			Telegram: config.TelegramConfigs{
				ReclaimDelay: parseDuration(getEnv("TELEGRAM_RECLAIM_DELAY", "15m")),
				BotToken:     getEnv("TELEGRAM_BOT_TOKEN", "telegram-bot-token"),
			},
			QuizMaxQuestions:               parseInt(getEnv("QUIZ_MAX_QUESTIONS", "10")),
			QuizMaxOptions:                 parseInt(getEnv("QUIZ_MAX_OPTIONS", "10")),
			InviteReclaimDelay:             parseDuration(getEnv("INVITE_RECLAIM_DELAY", "1m")),
			InviteProjectReclaimDelay:      parseDuration(getEnv("INVITE_PROJECT_RECLAIM_DELAY", "1m")),
			InviteProjectRequiredFollowers: parseInt(getEnv("INVITE_PROJECT_REQUIRED_FOLLOWERS", "10000")),
			InviteProjectRewardToken:       getEnv("INVITE_PROJECT_REWARD_TOKEN", "USDT"),
			InviteProjectRewardAmount:      parseFloat64(getEnv("INVITE_PROJECT_REWARD_AMOUNT", "50")),
		},
		Redis: config.RedisConfigs{
			Addr: getEnv("REDIS_ADDRESS", "localhost:6379"),
		},
		Kafka: config.KafkaConfigs{
			Addr: getEnv("KAFKA_ADDRESS", "localhost:9092"),
		},
		Game: config.GameConfigs{
			GameSaveFrequency: parseDuration(getEnv("GAME_SAVE_FREQUENCY", "10s")),
			MoveActionDelay:   parseDuration(getEnv("MOVING_ACTION_DELAY", "10ms")),
			InitActionDelay:   parseDuration(getEnv("INIT_ACTION_DELAY", "10s")),
			JoinActionDelay:   parseDuration(getEnv("JOIN_ACTION_DELAY", "10s")),
		},
		Cron: config.CronConfigs{
			ProjectTrendingInterval: parseDuration(getEnv("PROJECT_TRENDING_INTERVAL", "168h")),
		},
	}
}

func (app *App) newDatabase() *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       xcontext.Configs(app.ctx).Database.ConnectionString(), // data source name
		DefaultStringSize:         256,                                                   // default size for string fields
		DisableDatetimePrecision:  true,                                                  // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,                                                  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,                                                  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false,                                                 // auto configure based on currently MySQL version
	}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(parseDatabaseLogLevel(xcontext.Configs(app.ctx).Database.LogLevel)),
	})
	if err != nil {
		panic(err)
	}

	return db
}

func (app *App) migrateDB() {
	if err := entity.MigrateTable(app.ctx); err != nil {
		panic(err)
	}

	if err := entity.MigrateMySQL(app.ctx); err != nil {
		panic(err)
	}
}

func (app *App) loadStorage() {
	app.storage = storage.NewS3Storage(xcontext.Configs(app.ctx).Storage)
}

func (app *App) loadEndpoint() {
	app.twitterEndpoint = twitter.New(xcontext.Configs(app.ctx).Quest.Twitter)
	app.discordEndpoint = discord.New(xcontext.Configs(app.ctx).Quest.Dicord)
	app.telegramEndpoint = telegram.New(xcontext.Configs(app.ctx).Quest.Telegram)
}

func (app *App) loadRepos() {
	app.userRepo = repository.NewUserRepository()
	app.oauth2Repo = repository.NewOAuth2Repository()
	app.projectRepo = repository.NewProjectRepository()
	app.questRepo = repository.NewQuestRepository()
	app.categoryRepo = repository.NewCategoryRepository()
	app.collaboratorRepo = repository.NewCollaboratorRepository()
	app.claimedQuestRepo = repository.NewClaimedQuestRepository()
	app.participantRepo = repository.NewParticipantRepository()
	app.fileRepo = repository.NewFileRepository()
	app.apiKeyRepo = repository.NewAPIKeyRepository()
	app.refreshTokenRepo = repository.NewRefreshTokenRepository()
	app.userAggregateRepo = repository.NewUserAggregateRepository()
	app.gameRepo = repository.NewGameRepository()
	app.badgeRepo = repository.NewBadgeRepository()
	app.transactionRepo = repository.NewTransactionRepository()
}

func (app *App) loadBadgeManager() {
	app.badgeManager = badge.NewManager(
		app.badgeRepo,
		badge.NewSharpScoutBadgeScanner(app.participantRepo, []uint64{1, 2, 5, 10, 50}),
		badge.NewRainBowBadgeScanner(app.participantRepo, []uint64{3, 7, 14, 30, 50, 75, 125, 180, 250, 365}),
		badge.NewQuestWarriorBadgeScanner(app.userAggregateRepo, []uint64{3, 5, 10, 18, 30}),
	)
}

func (app *App) loadDomains() {
	cfg := xcontext.Configs(app.ctx)
	app.authDomain = domain.NewAuthDomain(app.userRepo, app.refreshTokenRepo, app.oauth2Repo,
		cfg.Auth.Google, cfg.Auth.Twitter, cfg.Auth.Discord)
	app.userDomain = domain.NewUserDomain(app.userRepo, app.oauth2Repo, app.participantRepo, app.badgeRepo,
		app.projectRepo, app.badgeManager, app.storage)
	app.projectDomain = domain.NewProjectDomain(app.projectRepo, app.collaboratorRepo, app.userRepo,
		app.discordEndpoint, app.storage)
	app.questDomain = domain.NewQuestDomain(app.questRepo, app.projectRepo, app.categoryRepo,
		app.collaboratorRepo, app.userRepo, app.claimedQuestRepo, app.oauth2Repo, app.transactionRepo,
		app.twitterEndpoint, app.discordEndpoint, app.telegramEndpoint)
	app.categoryDomain = domain.NewCategoryDomain(app.categoryRepo, app.projectRepo, app.collaboratorRepo,
		app.userRepo)
	app.collaboratorDomain = domain.NewCollaboratorDomain(app.projectRepo, app.collaboratorRepo, app.userRepo)
	app.claimedQuestDomain = domain.NewClaimedQuestDomain(app.claimedQuestRepo, app.questRepo,
		app.collaboratorRepo, app.participantRepo, app.oauth2Repo, app.userAggregateRepo, app.userRepo,
		app.projectRepo, app.transactionRepo, app.twitterEndpoint, app.discordEndpoint, app.telegramEndpoint,
		app.badgeManager)
	app.fileDomain = domain.NewFileDomain(app.storage, app.fileRepo)
	app.apiKeyDomain = domain.NewAPIKeyDomain(app.apiKeyRepo, app.collaboratorRepo, app.userRepo)
	app.gameProxyDomain = domain.NewGameProxyDomain(app.gameRepo, app.proxyRouter, app.publisher)
	app.statisticDomain = domain.NewStatisticDomain(app.userAggregateRepo, app.userRepo)
	app.gameDomain = domain.NewGameDomain(app.gameRepo, app.userRepo, app.fileRepo, app.storage, cfg.File)
	app.participantDomain = domain.NewParticipantDomain(app.collaboratorRepo, app.userRepo, app.participantRepo)
	app.transactionDomain = domain.NewTransactionDomain(app.transactionRepo)
}

func (app *App) loadPublisher() {
	app.publisher = kafka.NewPublisher(uuid.NewString(), []string{xcontext.Configs(app.ctx).Kafka.Addr})
}

func parseDuration(s string) time.Duration {
	duration, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}

	return duration
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return i
}

func parseFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}

	return f
}

func parseEnvAsInt(key string, def int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return def
	}

	return parseInt(value)
}

func parseDatabaseLogLevel(s string) gormlogger.LogLevel {
	switch s {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "warn":
		return gormlogger.Warn
	case "info":
		return gormlogger.Info
	}

	panic(fmt.Sprintf("invalid gorm log level %s", s))
}
