package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/domain"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/twitter"
	"github.com/questx-lab/backend/pkg/logger"
	"github.com/questx-lab/backend/pkg/pubsub"
	redisutil "github.com/questx-lab/backend/pkg/redis"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/storage"

	"github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type srv struct {
	app *cli.App

	authVerifier *middleware.AuthVerifier

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
	roomRepo          repository.RoomRepository
	userAggregateRepo repository.UserAggregateRepository

	userDomain         domain.UserDomain
	authDomain         domain.AuthDomain
	projectDomain      domain.ProjectDomain
	questDomain        domain.QuestDomain
	categoryDomain     domain.CategoryDomain
	collaboratorDomain domain.CollaboratorDomain
	claimedQuestDomain domain.ClaimedQuestDomain
	fileDomain         domain.FileDomain
	apiKeyDomain       domain.APIKeyDomain
	wsDomain           domain.WsDomain

	requestPublisher   pubsub.Publisher
	responseSubscriber pubsub.Subscriber
	statisticDomain    domain.StatisticDomain

	router *router.Router

	db          *gorm.DB
	redisClient *redis.Client

	configs *config.Configs

	server *http.Server

	storage         storage.Storage
	twitterEndpoint twitter.IEndpoint
	logger          logger.Logger
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
	accessTokenDuration, err := time.ParseDuration(getEnv("ACCESS_TOKEN_DURATION", "5m"))
	if err != nil {
		panic(err)
	}

	refreshTokenDuration, err := time.ParseDuration(getEnv("REFRESH_TOKEN_DURATION", "20m"))
	if err != nil {
		panic(err)
	}

	twitterReclaimDelay, err := time.ParseDuration(getEnv("TWITTER_RECLAIM_DELAY", "15m"))
	if err != nil {
		panic(err)
	}

	maxUploadSize, _ := strconv.Atoi(getEnv("MAX_UPLOAD_FILE", "2"))
	s.configs = &config.Configs{
		Env: getEnv("ENV", "local"),
		ApiServer: config.ServerConfigs{
			Host: getEnv("HOST", "localhost"),
			Port: getEnv("PORT", "8080"),
			Cert: getEnv("SERVER_CERT", "cert"),
			Key:  getEnv("SERVER_KEY", "key"),
		},
		Auth: config.AuthConfigs{
			TokenSecret: getEnv("TOKEN_SECRET", "token_secret"),
			AccessToken: config.TokenConfigs{
				Name:       "access_token",
				Expiration: accessTokenDuration,
			},
			RefreshToken: config.TokenConfigs{
				Name:       "refresh_token",
				Expiration: refreshTokenDuration,
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
				ReclaimDelay:      twitterReclaimDelay,
				AppAccessToken:    getEnv("TWITTER_APP_ACCESS_TOKEN", "app_access_token"),
				ConsumerAPIKey:    getEnv("TWITTER_CONSUMER_API_KEY", "consumer_key"),
				ConsumerAPISecret: getEnv("TWITTER_CONSUMER_API_SECRET", "comsumer_secret"),
				AccessToken:       getEnv("TWITTER_ACCESS_TOKEN", "access_token"),
				AccessTokenSecret: getEnv("TWITTER_ACCESS_TOKEN_SECRET", "access_token_secret"),
			},
		},
		WsProxyServer: config.ServerConfigs{
			Host: getEnv("HOST", "localhost"),
			Port: getEnv("PORT", "8081"),
			Cert: getEnv("SERVER_CERT", "cert"),
			Key:  getEnv("SERVER_KEY", "key"),
		},
		Redis: config.RedisConfigs{
			Addr: getEnv("REDIS_ADDRESS", "localhost:6379"),
		},
		Kafka: config.KafkaConfigs{
			Addr: getEnv("KAFKA_ADDRESS", "localhost:9092"),
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

	s.redisClient = redisutil.NewClient(s.configs.Redis.Addr)
}

func (s *srv) loadStorage() {
	s.storage = storage.NewS3Storage(&s.configs.Storage)
}

func (s *srv) loadEndpoint() {
	var err error
	s.twitterEndpoint, err = twitter.New(context.Background(), s.configs.Quest.Twitter)
	if err != nil {
		panic(err)
	}
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
	s.roomRepo = repository.NewRoomRepository()
	s.userAggregateRepo = repository.NewUserAggregateRepository()
}

func (s *srv) loadDomains() {
	s.authDomain = domain.NewAuthDomain(s.userRepo, s.refreshTokenRepo, s.oauth2Repo,
		s.configs.Auth.Google, s.configs.Auth.Twitter)
	s.userDomain = domain.NewUserDomain(s.userRepo, s.participantRepo)
	s.projectDomain = domain.NewProjectDomain(s.projectRepo, s.collaboratorRepo)
	s.questDomain = domain.NewQuestDomain(s.questRepo, s.projectRepo, s.categoryRepo,
		s.collaboratorRepo, s.twitterEndpoint)
	s.categoryDomain = domain.NewCategoryDomain(s.categoryRepo, s.projectRepo, s.collaboratorRepo)
	s.collaboratorDomain = domain.NewCollaboratorDomain(s.projectRepo, s.collaboratorRepo, s.userRepo)
	s.claimedQuestDomain = domain.NewClaimedQuestDomain(s.claimedQuestRepo, s.questRepo,
		s.collaboratorRepo, s.participantRepo, s.oauth2Repo, s.userAggregateRepo, s.twitterEndpoint)
	s.fileDomain = domain.NewFileDomain(s.storage, s.fileRepo, s.configs.File)
	s.apiKeyDomain = domain.NewAPIKeyDomain(s.apiKeyRepo, s.collaboratorRepo)
	s.wsDomain = domain.NewWsDomain(s.roomRepo, s.authVerifier, s.requestPublisher)
	s.statisticDomain = domain.NewStatisticDomain(s.userAggregateRepo)
}

func (s *srv) loadAuthVerifier() {
	s.authVerifier = middleware.NewAuthVerifier().WithAccessToken()
}
