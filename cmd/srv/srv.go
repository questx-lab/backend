package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/domain"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/middleware"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/pkg/router"
	"gorm.io/gorm"

	"gorm.io/driver/mysql"
)

type srv struct {
	userRepo    repository.UserRepository
	oauth2Repo  repository.OAuth2Repository
	projectRepo repository.ProjectRepository

	userDomain       domain.UserDomain
	oauth2Domain     domain.OAuth2Domain
	walletAuthDomain domain.WalletAuthDomain
	projectDomain    domain.ProjectDomain

	router *router.Router

	db *gorm.DB

	configs *config.Configs

	server *http.Server
}

func (s *srv) loadConfig() {
	tokenDuration, err := time.ParseDuration(os.Getenv("TOKEN_DURATION"))
	if err != nil {
		panic(err)
	}

	s.configs = &config.Configs{
		Server: config.ServerConfigs{
			Host: os.Getenv("HOST"),
			Port: os.Getenv("PORT"),
			Cert: os.Getenv("SERVER_CERT"),
			Key:  os.Getenv("SERVER_KEY"),
		},
		Auth: config.AuthConfigs{
			AccessTokenName: "questx_token",
			Google: config.OAuth2Config{
				Name:         "google",
				Issuer:       "https://accounts.google.com",
				ClientID:     os.Getenv("OAUTH2_GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("OAUTH2_GOOGLE_CLIENT_SECRET"),
				IDField:      "email",
			},
		},
		Database: config.DatabaseConfigs{
			Host:     os.Getenv("MYSQL_HOST"),
			Port:     os.Getenv("MYSQL_PORT"),
			User:     os.Getenv("MYSQL_USER"),
			Password: os.Getenv("MYSQL_PASSWORD"),
			Database: os.Getenv("MYSQL_DATABASE"),
		},
		Token: config.TokenConfigs{
			Secret:     os.Getenv("TOKEN_SECRET"),
			Expiration: tokenDuration,
		},
		Session: config.SessionConfigs{
			Secret: os.Getenv("AUTH_SESSION_SECRET"),
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
	s.db.AutoMigrate(&entity.User{}, &entity.OAuth2{})
}

func (s *srv) loadRepos() {
	s.userRepo = repository.NewUserRepository(s.db)
	s.oauth2Repo = repository.NewOAuth2Repository(s.db)
	s.projectRepo = repository.NewProjectRepository(s.db)
}

func (s *srv) loadDomains() {
	authenticators := setupOAuth2(s.configs.Auth.Google)
	s.oauth2Domain = domain.NewOAuth2Domain(s.userRepo, s.oauth2Repo, authenticators)
	s.walletAuthDomain = domain.NewWalletAuthDomain(s.userRepo)
	s.userDomain = domain.NewUserDomain(s.userRepo)
	s.projectDomain = domain.NewProjectDomain(s.projectRepo)
}

func (s *srv) loadRouter() {
	s.router = router.New(*s.configs)

	gob.Register(map[string]interface{}{})
	store := cookie.NewStore([]byte(s.configs.Session.Secret))
	s.router.Inner.Use(sessions.Sessions(s.configs.Session.Name, store))

	router.GET(s.router, "/oauth2/login", s.oauth2Domain.Login)
	router.GET(s.router, "/oauth2/callback", s.oauth2Domain.Callback)
	router.GET(s.router, "/wallet/login", s.walletAuthDomain.Login)
	router.GET(s.router, "/wallet/verify", s.walletAuthDomain.Verify)

	needAuthRouter := s.router.Branch()
	needAuthRouter.Use(middleware.Authenticate())
	{
		router.GET(needAuthRouter, "/getUser", s.userDomain.GetUser)
		router.GET(s.router, "/getListProject", s.projectDomain.GetList)
	}
}

func (s *srv) startServer() {
	s.router.Static("/static", "./web")
	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", s.configs.Server.Host, s.configs.Server.Port),
		Handler: s.router.Handler(),
	}

	log.Printf("Starting server on port: %s\n", s.configs.Server.Port)
	if err := s.server.ListenAndServeTLS(s.configs.Server.Cert, s.configs.Server.Key); err != nil {
		panic(err)
	}
}

func setupOAuth2(configs ...config.OAuth2Config) []authenticator.OAuth2 {
	authenticators := make([]authenticator.OAuth2, len(configs))
	for i, cfg := range configs {
		authenticator, err := authenticator.NewOAuth2(context.Background(), cfg)
		if err != nil {
			panic(err)
		}
		authenticators[i] = authenticator
	}

	return authenticators
}
