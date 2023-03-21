package main

import (
	"context"
	"fmt"
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
	"gorm.io/gorm"
	"github.com/questx-lab/backend/utils/token"

	"gorm.io/driver/mysql"
)

type controller interface {
	Register(mux *http.ServeMux)
}

type srv struct {
	userRepo   repository.UserRepository
	oauth2Repo repository.OAuth2Repository
	
	tknGenerator token.Generator

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
		Database: config.DatabaseConfig{
			DSN: os.Getenv("DSN"),
		},
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
	}
}

func (s *srv) loadDatabase() {
	var err error
	s.db, err = gorm.Open(mysql.Open(s.configs.Database.DSN), &gorm.Config{})
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
	var authenticators []authenticator.OAuth2
	authenticators = setupOAuth2(
		s.configs.Auth.Google,
		s.configs.Auth.Twitter,
	)
	s.oauth2Domain = domain.NewOAuth2Domain(s.userRepo, s.oauth2Repo, authenticators, s.configs.Auth)
	s.walletAuthDomain = domain.NewWalletAuthDomain(s.userRepo, s.configs.Auth)
	s.userDomain = domain.NewUserDomain(s.userRepo)
	s.authDomain = domain.NewAuthDomain(s.userRepo)
	s.projectDomain = domain.NewProjectDomain(s.projectRepo)
}

func (s *srv) loadControllers() {
	controllers := []controller{
		&api.Endpoint[model.OAuth2LoginRequest, model.OAuth2LoginResponse]{
			Path:   "/oauth2/login",
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
			Path:   "/get_user",
			Method: http.MethodGet,
			Handle: s.userDomain.GetUser,

		&api.Endpoint[model.RegisterRequest, model.RegisterResponse]{
			Path:   "/auth/register",
			Method: http.MethodPost,
			Handle: s.authDomain.Register,
		},

		&api.Endpoint[model.CreateProjectRequest, model.CreateProjectResponse]{
			Path:   "/projects",
			Method: http.MethodPost,
			Handle: s.projectDomain.CreateProject,
			Before: []api.Handler{
				api.ImportUserIDToContext(s.tknGenerator),
			},
		},
	}

	for _, e := range controllers {
		e.Register(s.mux)
	}
}

func (s *srv) startServer() {
	fmt.Println("Starting server")
	s.mux.Handle("/", http.FileServer(http.Dir("./web")))
	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", s.configs.Server.Host, s.configs.Server.Port),
		Handler: s.mux,
	}

	if err := s.server.ListenAndServeTLS(s.configs.Server.Cert, s.configs.Server.Key); err != nil {
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
