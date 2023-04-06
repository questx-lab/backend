package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/domain"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/storage"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type srv struct {
	userRepo         repository.UserRepository
	oauth2Repo       repository.OAuth2Repository
	projectRepo      repository.ProjectRepository
	questRepo        repository.QuestRepository
	categoryRepo     repository.CategoryRepository
	collaboratorRepo repository.CollaboratorRepository
	claimedQuestRepo repository.ClaimedQuestRepository
	participantRepo  repository.ParticipantRepository
	fileRepo         repository.FileRepository
	apiKeyRepo       repository.APIKeyRepository

	userDomain         domain.UserDomain
	oauth2Domain       domain.OAuth2Domain
	walletAuthDomain   domain.WalletAuthDomain
	projectDomain      domain.ProjectDomain
	questDomain        domain.QuestDomain
	categoryDomain     domain.CategoryDomain
	collaboratorDomain domain.CollaboratorDomain
	claimedQuestDomain domain.ClaimedQuestDomain
	fileDomain         domain.FileDomain
	apiKeyDomain       domain.APIKeyDomain

	router *router.Router

	db *gorm.DB

	configs *config.Configs

	server *http.Server

	storage storage.Storage
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func (s *srv) loadConfig() {
	tokenDuration, err := time.ParseDuration(getEnv("TOKEN_DURATION", "5m"))
	if err != nil {
		panic(err)
	}

	s.configs = &config.Configs{
		Env: getEnv("ENV", "local"),
		Server: config.ServerConfigs{
			Host: getEnv("HOST", "localhost"),
			Port: getEnv("PORT", "8080"),
			Cert: getEnv("SERVER_CERT", "cert"),
			Key:  getEnv("SERVER_KEY", "key"),
		},
		Auth: config.AuthConfigs{
			AccessTokenName: "questx_token",
			CallbackURL:     os.Getenv("AUTH_CALLBACK_URL"),
			Google: config.OAuth2Config{
				Name:         "google",
				Issuer:       "https://accounts.google.com",
				ClientID:     getEnv("OAUTH2_GOOGLE_CLIENT_ID", "client_id"),
				ClientSecret: getEnv("OAUTH2_GOOGLE_CLIENT_SECRET", "secret_id"),
				IDField:      "email",
			},
		},
		Database: config.DatabaseConfigs{
			Host:     getEnv("MYSQL_HOST", "mysql"),
			Port:     getEnv("MYSQL_PORT", "3306"),
			User:     getEnv("MYSQL_USER", "mysql"),
			Password: getEnv("MYSQL_PASSWORD", "mysql"),
			Database: getEnv("MYSQL_DATABASE", "questx"),
		},
		Token: config.TokenConfigs{
			Secret:     getEnv("TOKEN_SECRET", "token_secret"),
			Expiration: tokenDuration,
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
}

func (s *srv) loadStorage() {
	s.storage = storage.NewS3Storage(&s.configs.Storage)
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
}

func (s *srv) loadDomains() {
	oauth2Configs := setupOAuth2(*s.configs, s.configs.Auth.Google)
	s.oauth2Domain = domain.NewOAuth2Domain(s.userRepo, s.oauth2Repo, oauth2Configs)
	s.walletAuthDomain = domain.NewWalletAuthDomain(s.userRepo)
	s.userDomain = domain.NewUserDomain(s.userRepo, s.participantRepo)
	s.projectDomain = domain.NewProjectDomain(s.projectRepo, s.collaboratorRepo)
	s.questDomain = domain.NewQuestDomain(s.questRepo, s.projectRepo, s.categoryRepo, s.collaboratorRepo)
	s.categoryDomain = domain.NewCategoryDomain(s.categoryRepo, s.projectRepo, s.collaboratorRepo)
	s.collaboratorDomain = domain.NewCollaboratorDomain(s.projectRepo, s.collaboratorRepo, s.userRepo)
	s.claimedQuestDomain = domain.NewClaimedQuestDomain(
		s.claimedQuestRepo, s.questRepo, s.collaboratorRepo, s.participantRepo)
	s.fileDomain = domain.NewFileDomain(s.storage, s.fileRepo)
	s.apiKeyDomain = domain.NewAPIKeyDomain(s.apiKeyRepo, s.collaboratorRepo)
}

func (s *srv) loadRouter() {
	s.router = router.New(s.db, *s.configs)
	s.router.Static("/", "./web")
	s.router.AddCloser(middleware.Logger())

	// Auth API
	authRouter := s.router.Branch()
	authRouter.After(middleware.HandleSaveSession())
	authRouter.After(middleware.HandleSetAccessToken())
	authRouter.After(middleware.HandleRedirect())
	{
		router.GET(authRouter, "/oauth2/login", s.oauth2Domain.Login)
		router.GET(authRouter, "/oauth2/callback", s.oauth2Domain.Callback)
		router.GET(authRouter, "/wallet/login", s.walletAuthDomain.Login)
		router.GET(authRouter, "/wallet/verify", s.walletAuthDomain.Verify)
	}

	// These following APIs need authentication with only Access Token.
	onlyTokenAuthRouter := s.router.Branch()
	authVerifier := middleware.NewAuthVerifier().WithAccessToken()
	onlyTokenAuthRouter.Before(authVerifier.Middleware())
	{
		// User API
		router.GET(onlyTokenAuthRouter, "/getUser", s.userDomain.GetUser)
		router.GET(onlyTokenAuthRouter, "/getPoints", s.userDomain.GetPoints)
		router.POST(onlyTokenAuthRouter, "/joinProject", s.userDomain.JoinProject)

		// Project API
		router.POST(onlyTokenAuthRouter, "/createProject", s.projectDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateProjectByID", s.projectDomain.UpdateByID)
		router.POST(onlyTokenAuthRouter, "/deleteProjectByID", s.projectDomain.DeleteByID)

		// API-Key API
		router.POST(onlyTokenAuthRouter, "/generateAPIKey", s.apiKeyDomain.Generate)
		router.POST(onlyTokenAuthRouter, "/regenerateAPIKey", s.apiKeyDomain.Regenerate)
		router.POST(onlyTokenAuthRouter, "/revokeAPIKey", s.apiKeyDomain.Revoke)

		// Collaborator API
		router.GET(onlyTokenAuthRouter, "/getListCategory", s.categoryDomain.GetList)

		// Quest API
		router.POST(onlyTokenAuthRouter, "/createQuest", s.questDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateQuest", s.questDomain.Update)

		// Category API
		router.POST(onlyTokenAuthRouter, "/createCategory", s.categoryDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateCategoryByID", s.categoryDomain.UpdateByID)
		router.POST(onlyTokenAuthRouter, "/deleteCategoryByID", s.categoryDomain.DeleteByID)

		// Collaborator API
		router.GET(onlyTokenAuthRouter, "/getListCollaborator", s.collaboratorDomain.GetList)
		router.POST(onlyTokenAuthRouter, "/createCollaborator", s.collaboratorDomain.Create)
		router.POST(onlyTokenAuthRouter, "/updateCollaboratorByID", s.collaboratorDomain.UpdateRole)
		router.POST(onlyTokenAuthRouter, "/deleteCollaboratorByID", s.collaboratorDomain.Delete)

		// Claimed Quest API
		router.GET(onlyTokenAuthRouter, "/getClaimedQuest", s.claimedQuestDomain.Get)
		router.GET(onlyTokenAuthRouter, "/getListClaimedQuest", s.claimedQuestDomain.GetList)
		router.POST(onlyTokenAuthRouter, "/claim", s.claimedQuestDomain.Claim)

		router.POST(onlyTokenAuthRouter, "/uploadImage", s.fileDomain.UploadImage)
	}

	// These following APIs support authentication with both Access Token and API Key.
	tokenAndKeyAuthRouter := s.router.Branch()
	authVerifier = middleware.NewAuthVerifier().WithAccessToken().WithAPIKey(s.apiKeyRepo)
	tokenAndKeyAuthRouter.Before(authVerifier.Middleware())
	{
		router.GET(tokenAndKeyAuthRouter, "/getClaimedQuest", s.claimedQuestDomain.Get)
		router.GET(tokenAndKeyAuthRouter, "/getListClaimedQuest", s.claimedQuestDomain.GetList)
		router.GET(tokenAndKeyAuthRouter, "/getPendingClaimedQuestList", s.claimedQuestDomain.GetPendingList)
		router.POST(tokenAndKeyAuthRouter, "/reviewClaimedQuest", s.claimedQuestDomain.ReviewClaimedQuest)
	}

	// Public API.
	router.GET(s.router, "/getQuest", s.questDomain.Get)
	router.GET(s.router, "/getListQuest", s.questDomain.GetList)
	router.GET(s.router, "/getListProject", s.projectDomain.GetList)
	router.GET(s.router, "/getProjectByID", s.projectDomain.GetByID)
}

func (s *srv) startServer() {
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.configs.Server.Port),
		Handler: s.router.Handler(),
	}

	fmt.Printf("Starting server on port: %s\n", s.configs.Server.Port)
	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
	fmt.Printf("server stop")
}

func setupOAuth2(cfg config.Configs, oauth2Cfgs ...config.OAuth2Config) []authenticator.IOAuth2Config {
	oauth2Configs := make([]authenticator.IOAuth2Config, len(oauth2Cfgs))
	for i, oauth2Cfg := range oauth2Cfgs {
		authenticator, err := authenticator.NewOAuth2Config(context.Background(), cfg, oauth2Cfg)
		if err != nil {
			panic(err)
		}
		oauth2Configs[i] = authenticator
	}

	return oauth2Configs
}
