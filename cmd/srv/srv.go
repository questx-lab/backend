package main

import (
	"context"
	"fmt"
	"log"
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

	userDomain         domain.UserDomain
	oauth2Domain       domain.OAuth2Domain
	walletAuthDomain   domain.WalletAuthDomain
	projectDomain      domain.ProjectDomain
	questDomain        domain.QuestDomain
	categoryDomain     domain.CategoryDomain
	collaboratorDomain domain.CollaboratorDomain

	router *router.Router

	db *gorm.DB

	configs *config.Configs

	server *http.Server
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

func (s *srv) loadRepos() {
	s.userRepo = repository.NewUserRepository(s.db)
	s.oauth2Repo = repository.NewOAuth2Repository(s.db)
	s.projectRepo = repository.NewProjectRepository(s.db)
	s.questRepo = repository.NewQuestRepository(s.db)
	s.categoryRepo = repository.NewCategoryRepository(s.db)
	s.collaboratorRepo = repository.NewCollaboratorRepository(s.db)
}

func (s *srv) loadDomains() {
	oauth2Configs := setupOAuth2(s.configs.Auth.Google)
	s.oauth2Domain = domain.NewOAuth2Domain(s.userRepo, s.oauth2Repo, oauth2Configs)
	s.walletAuthDomain = domain.NewWalletAuthDomain(s.userRepo)
	s.userDomain = domain.NewUserDomain(s.userRepo)
	s.projectDomain = domain.NewProjectDomain(s.projectRepo, s.collaboratorRepo)
	s.questDomain = domain.NewQuestDomain(s.questRepo, s.projectRepo, s.categoryRepo)
	s.categoryDomain = domain.NewCategoryDomain(s.categoryRepo, s.projectRepo, s.collaboratorRepo)
	s.collaboratorDomain = domain.NewCollaboratorDomain(s.projectRepo, s.collaboratorRepo, s.userRepo)
}

func (s *srv) loadRouter() {
	s.router = router.New(*s.configs)
	s.router.Static("/", "./web")
	s.router.AddCloser(middleware.Logger())

	authRouter := s.router.Branch()
	authRouter.After(middleware.HandleSaveSession())
	authRouter.After(middleware.HandleSetAccessToken())
	authRouter.After(middleware.HandleRedirect())

	// Auth API
	{
		router.GET(authRouter, "/oauth2/login", s.oauth2Domain.Login)
		router.GET(authRouter, "/oauth2/callback", s.oauth2Domain.Callback)
		router.GET(authRouter, "/wallet/login", s.walletAuthDomain.Login)
		router.GET(authRouter, "/wallet/verify", s.walletAuthDomain.Verify)
	}

	needAuthRouter := s.router.Branch()
	needAuthRouter.Before(middleware.Authenticate())
	{
		// User API
		router.POST(needAuthRouter, "/getUser", s.userDomain.GetUser)

		// Project API
		router.POST(needAuthRouter, "/createProject", s.projectDomain.Create)
		router.POST(needAuthRouter, "/updateProjectByID", s.projectDomain.UpdateByID)
		router.POST(needAuthRouter, "/deleteProjectByID", s.projectDomain.DeleteByID)

		// Quest API
		router.POST(needAuthRouter, "/createQuest", s.questDomain.Create)

		// Category API
		router.POST(needAuthRouter, "/createCategory", s.categoryDomain.Create)
		router.POST(needAuthRouter, "/updateCategoryByID", s.categoryDomain.UpdateByID)
		router.POST(needAuthRouter, "/deleteCategoryByID", s.categoryDomain.DeleteByID)

		// Collaborator API
		router.POST(needAuthRouter, "/createCollaborator", s.collaboratorDomain.Create)
		router.POST(needAuthRouter, "/updateCollaboratorByID", s.collaboratorDomain.UpdateRole)
		router.POST(needAuthRouter, "/deleteCollaboratorByID", s.collaboratorDomain.Delete)
	}

	// For get by id, get list
	router.GET(s.router, "/getQuest", s.questDomain.Get)
	router.GET(s.router, "/getQuests", s.questDomain.GetList)
	router.GET(s.router, "/getListCategory", s.categoryDomain.GetList)
	router.GET(s.router, "/getListCollaborator", s.collaboratorDomain.GetList)
}

func (s *srv) startServer() {
	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", s.configs.Server.Host, s.configs.Server.Port),
		Handler: s.router.Handler(),
	}

	log.Printf("Starting server on port: %s\n", s.configs.Server.Port)
	if s.configs.Env == "local" {
		if err := s.server.ListenAndServe(); err != nil {
			panic(err)
		}
	} else {
		if err := s.server.ListenAndServeTLS(s.configs.Server.Cert, s.configs.Server.Key); err != nil {
			panic(err)
		}
	}
}

func setupOAuth2(configs ...config.OAuth2Config) []authenticator.IOAuth2Config {
	oauth2Configs := make([]authenticator.IOAuth2Config, len(configs))
	for i, cfg := range configs {
		authenticator, err := authenticator.NewOAuth2(context.Background(), cfg)
		if err != nil {
			panic(err)
		}
		oauth2Configs[i] = authenticator
	}

	return oauth2Configs
}
