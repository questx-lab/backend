package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/domain"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/authenticator"
	"github.com/questx-lab/backend/utils/token"
	"gorm.io/gorm"

	"gorm.io/driver/mysql"
)

type controller interface {
	Register(mux *http.ServeMux)
}

type srv struct {
	userRepo   repository.UserRepository
	oauth2Repo repository.OAuth2Repository

	accessTokenGenerator  token.Generator
	refreshTokenGenerator token.Generator

	controllers []controller

	userDomain       domain.UserDomain
	oauth2Domain     domain.OAuth2Domain
	walletAuthDomain domain.WalletAuthDomain

	mux *http.ServeMux

	db *gorm.DB

	configs *config.Configs

	server *http.Server
}

func (s *srv) loadMux() {
	s.mux = http.NewServeMux()
}

func (s *srv) loadConfig() {
	godotenv.Load(".env")

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
			TokenSecret:     os.Getenv("TOKEN_SECRET"),
			TokenExpiration: tokenDuration,
			AccessTokenName: "questx_token",
			SessionSecret:   os.Getenv("AUTH_SESSION_SECRET"),
			Google: config.OAuth2Config{
				Name:         "google",
				Issuer:       "https://accounts.google.com",
				ClientID:     os.Getenv("OAUTH2_GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("OAUTH2_GOOGLE_CLIENT_SECRET"),
				IDField:      "email",
			},
			Twitter: config.OAuth2Config{
				Name:         "twitter",
				Issuer:       "https://twitter.com",
				ClientID:     os.Getenv("OAUTH2_TWITTER_CLIENT_ID"),
				ClientSecret: os.Getenv("OAUTH2_TWITTER_CLIENT_SECRET"),
				IDField:      "???",
			},
		},
		DB: &config.Database{
			Host:     os.Getenv("MYSQL_HOST"),
			Port:     os.Getenv("MYSQL_PORT"),
			User:     os.Getenv("MYSQL_USER"),
			Password: os.Getenv("MYSQL_PASSWORD"),
			Database: os.Getenv("MYSQL_DATABASE"),
		},
		Port: os.Getenv("PORT"),
	}
}

func (s *srv) loadDatabase() {
	var err error
	s.db, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       s.configs.DB.ConnectionString(), // data source name
		DefaultStringSize:         256,                             // default size for string fields
		DisableDatetimePrecision:  true,                            // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,                            // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,                            // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false,                           // auto configure based on currently MySQL version
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	s.db.AutoMigrate(&entity.User{}, &entity.OAuth2{})
}

func (s *srv) loadRepos() {
	s.userRepo = repository.NewUserRepository(s.db)
	s.oauth2Repo = repository.NewOAuth2Repository(s.db)
}

func (s *srv) loadDomains() {
	var authenticators []authenticator.OAuth2
	authenticators = setupOAuth2(
		s.configs.Auth.Google,
		s.configs.Auth.Twitter,
	)
	s.oauth2Domain = domain.NewOAuth2Domain(s.userRepo, s.oauth2Repo, authenticators, s.accessTokenGenerator, s.configs.Auth)
	s.walletAuthDomain = domain.NewWalletAuthDomain(s.userRepo, s.configs.Auth)
	s.userDomain = domain.NewUserDomain(s.userRepo)
}

func (s *srv) loadControllers() {
	authGroup := &api.Group{
		Path:   "oauth2/auth",
		Before: []api.Handler{api.Logger},
		After:  []api.Handler{api.Close},
	}

	s.controllers = []controller{
		&api.Endpoint[model.OAuth2LoginRequest, model.OAuth2LoginResponse]{
			Group:  authGroup,
			Path:   "/login",
			Method: http.MethodGet,
			Handle: s.oauth2Domain.Login,
		},
		&api.Endpoint[model.OAuth2CallbackRequest, model.OAuth2CallbackResponse]{
			Path:   "/oauth2/callback",
			Method: http.MethodGet,
			Handle: s.oauth2Domain.Callback,
		},
		&api.Endpoint[model.WalletLoginRequest, model.WalletLoginResponse]{
			Path:   "/wallet/login",
			Method: http.MethodGet,
			Handle: s.walletAuthDomain.Login,
		},
		&api.Endpoint[model.WalletVerifyRequest, model.WalletVerifyResponse]{
			Path:   "/wallet/verify",
			Method: http.MethodGet,
			Handle: s.walletAuthDomain.Verify,
		},
		&api.Endpoint[model.GetUserRequest, model.GetUserResponse]{
			Path:   "/getUser",
			Method: http.MethodGet,
			Handle: s.userDomain.GetUser,
		},
	}
	for _, c := range s.controllers {
		c.Register(s.mux)
	}
}

func (s *srv) startServer() {
	s.mux.Handle("/", http.FileServer(http.Dir("./web")))
	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", s.configs.Server.Host, s.configs.Server.Port),
		Handler: s.mux,
	}

	log.Printf("Starting server on port: %s\n", s.configs.Port)
	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func setupOAuth2(configs ...config.OAuth2Config) []authenticator.OAuth2 {
	authenticators := make([]authenticator.OAuth2, 0, len(configs))
	for i, cfg := range configs {
		authenticator, err := authenticator.NewOAuth2(
			context.Background(),
			cfg.Name, cfg.Issuer,
			cfg.ClientID,
			cfg.ClientSecret,
			cfg.IDField,
		)
		if err != nil {
			panic(err)
		}

		authenticators[i] = authenticator
	}

	return authenticators
}

func (s *srv) loadAuthenticator() {
	s.accessTokenGenerator = token.NewJWTGenerator("access_token", &s.configs.TknConfigs)
	s.refreshTokenGenerator = token.NewJWTGenerator("refresh_token", &s.configs.TknConfigs)
}
