package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
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
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/pubsub"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/storage"

	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type srv struct {
	app *cli.App

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

	publisher   pubsub.Publisher
	proxyRouter gameproxy.Router

	db *gorm.DB

	server  *http.Server
	router  *router.Router
	configs *config.Configs

	storage storage.Storage

	logger logger.Logger

	badgeManager     *badge.Manager
	twitterEndpoint  twitter.IEndpoint
	discordEndpoint  discord.IEndpoint
	telegramEndpoint telegram.IEndpoint
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func (s *srv) loadLogger() {
	s.logger = logger.NewLogger()
}

func (s *srv) loadConfig() {
	maxUploadSize, _ := strconv.Atoi(getEnv("MAX_UPLOAD_FILE", "2"))
	s.configs = &config.Configs{
		Env: getEnv("ENV", "local"),
		ApiServer: config.APIServerConfigs{
			MaxLimit:     parseInt(getEnv("API_MAX_LIMIT", "50")),
			DefaultLimit: parseInt(getEnv("API_DEFAULT_LIMIT", "1")),
			ServerConfigs: config.ServerConfigs{
				Host: getEnv("API_HOST", "localhost"),
				Port: getEnv("API_PORT", "8080"),
			},
		},
		GameProxyServer: config.ServerConfigs{
			Host: getEnv("GAME_PROXY_HOST", "localhost"),
			Port: getEnv("GAME_PROXY_PORT", "8081"),
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
			MaxSize: maxUploadSize,
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
				BotToken: getEnv("DISCORD_BOT_TOKEN", "discord_bot_token"),
				BotID:    getEnv("DISCORD_BOT_ID", "discord_bot_id"),
			},
			Telegram: config.TelegramConfigs{
				BotToken: getEnv("TELEGRAM_BOT_TOKEN", "telegram-bot-token"),
			},
			Quiz: config.QuizConfigs{
				MaxQuestions: parseInt(getEnv("QUIZ_MAX_QUESTIONS", "10")),
				MaxOptions:   parseInt(getEnv("QUIZ_MAX_OPTIONS", "10")),
			},
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
	}
}

func (s *srv) loadDatabase() {
	var err error
	s.db, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       s.configs.Database.ConnectionString(), // data source name
		DefaultStringSize:         256,                                   // default size for string fields
		DisableDatetimePrecision:  true,                                  // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,                                  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,                                  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false,                                 // auto configure based on currently MySQL version
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err := entity.MigrateTable(s.db); err != nil {
		panic(err)
	}

	entity.MigrateMySQL(s.db, s.logger)
}

func (s *srv) loadStorage() {
	s.storage = storage.NewS3Storage(&s.configs.Storage)
}
func (s *srv) loadEndpoint() {
	s.twitterEndpoint = twitter.New(context.Background(), s.configs.Quest.Twitter)
	s.discordEndpoint = discord.New(context.Background(), s.configs.Quest.Dicord)
	s.telegramEndpoint = telegram.New(context.Background(), s.configs.Quest.Telegram)
}

func (s *srv) loadRepos() {
	s.userRepo = repository.NewUserRepository()
	s.oauth2Repo = repository.NewOAuth2Repository()
	s.projectRepo = repository.NewProjectRepository()
	s.questRepo = repository.NewQuestRepository()
	s.categoryRepo = repository.NewCategoryRepository()
	s.collaboratorRepo = repository.NewCollaboratorRepository()
	s.claimedQuestRepo = repository.NewClaimedQuestRepository()
	s.participantRepo = repository.NewParticipantRepository()
	s.fileRepo = repository.NewFileRepository()
	s.apiKeyRepo = repository.NewAPIKeyRepository()
	s.refreshTokenRepo = repository.NewRefreshTokenRepository()
	s.userAggregateRepo = repository.NewUserAggregateRepository()
	s.gameRepo = repository.NewGameRepository()
	s.badgeRepo = repository.NewBadgeRepository()
}

func (s *srv) loadBadgeManager() {
	s.badgeManager = badge.NewManager(
		s.badgeRepo,
		badge.NewSharpScoutBadgeScanner(s.participantRepo, []uint64{1, 2, 5, 10, 50}),
	)
}

func (s *srv) loadDomains() {
	s.authDomain = domain.NewAuthDomain(s.userRepo, s.refreshTokenRepo, s.oauth2Repo,
		s.configs.Auth.Google, s.configs.Auth.Twitter, s.configs.Auth.Discord)
	s.userDomain = domain.NewUserDomain(s.userRepo, s.oauth2Repo, s.participantRepo,
		s.badgeRepo, s.badgeManager)
	s.projectDomain = domain.NewProjectDomain(s.projectRepo, s.collaboratorRepo, s.userRepo, s.discordEndpoint)
	s.questDomain = domain.NewQuestDomain(s.questRepo, s.projectRepo, s.categoryRepo,
		s.collaboratorRepo, s.userRepo, s.claimedQuestRepo, s.oauth2Repo,
		s.twitterEndpoint, s.discordEndpoint, s.telegramEndpoint)
	s.categoryDomain = domain.NewCategoryDomain(s.categoryRepo, s.projectRepo, s.collaboratorRepo, s.userRepo)
	s.collaboratorDomain = domain.NewCollaboratorDomain(s.projectRepo, s.collaboratorRepo, s.userRepo)
	s.claimedQuestDomain = domain.NewClaimedQuestDomain(s.claimedQuestRepo, s.questRepo,
		s.collaboratorRepo, s.participantRepo, s.oauth2Repo, s.userAggregateRepo, s.userRepo, s.projectRepo,
		s.twitterEndpoint, s.discordEndpoint, s.telegramEndpoint)
	s.fileDomain = domain.NewFileDomain(s.storage, s.fileRepo, s.configs.File)
	s.apiKeyDomain = domain.NewAPIKeyDomain(s.apiKeyRepo, s.collaboratorRepo, s.userRepo)
	s.gameProxyDomain = domain.NewGameProxyDomain(s.gameRepo, s.proxyRouter, s.publisher)
	s.statisticDomain = domain.NewStatisticDomain(s.userAggregateRepo)
	s.gameDomain = domain.NewGameDomain(s.gameRepo, s.userRepo, s.fileRepo, s.storage, s.configs.File)
	s.participantDomain = domain.NewParticipantDomain(s.collaboratorRepo, s.userRepo, s.participantRepo)
}

func (s *srv) loadPublisher() {
	s.publisher = kafka.NewPublisher(uuid.NewString(), []string{s.configs.Kafka.Addr})
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
