package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/internal/domain"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/authenticator"

	_ "github.com/go-sql-driver/mysql"
)

type controller interface {
	Register(mux *http.ServeMux)
}

type srv struct {
	controllers []controller

	userRepo   repository.UserRepository
	oauth2Repo repository.OAuth2Repository

	userDomain   domain.UserDomain
	oauth2Domain domain.OAuth2Domain

	mux *http.ServeMux

	db *sql.DB

	configs *Configs

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

	s.configs = &Configs{
		DBConnection: os.Getenv("DB_CONNECTION"),
		Server: ServerConfigs{
			Address: os.Getenv("ADDRESS"),
			Port:    os.Getenv("PORT"),
			Cert:    os.Getenv("SERVER_CERT"),
			Key:     os.Getenv("SERVER_KEY"),
		},
		SessionSecret: os.Getenv("SESSION_SECRET"),
		Auth: AuthConfigs{
			TokenSecret:     os.Getenv("TOKEN_SECRET"),
			TokenExpiration: tokenDuration,
			Google: OAuth2Config{
				Name:         "google",
				Issuer:       "https://accounts.google.com",
				ClientID:     os.Getenv("OAUTH2_GOOGLE_CLIENT_ID"),
				ClientSecret: os.Getenv("OAUTH2_GOOGLE_CLIENT_SECRET"),
			},
			Tweeter: OAuth2Config{
				Name:         "tweeter",
				Issuer:       "https://tweeter.com",
				ClientID:     os.Getenv("OAUTH2_TWEETER_CLIENT_ID"),
				ClientSecret: os.Getenv("OAUTH2_TWEETER_CLIENT_SECRET"),
			},
		},
	}
}

func (s *srv) loadDatabase() {
	var err error
	s.db, err = sql.Open("mysql", s.configs.DBConnection)
	if err != nil {
		panic(err)
	}
}

func (s *srv) loadRepos() {
	s.userRepo = repository.NewUserRepository(s.db)
	s.oauth2Repo = repository.NewOAuth2Repository(s.db)
}

func (s *srv) loadDomains() {
	s.userDomain = domain.NewUserDomain(s.userRepo)

	var authenticators []authenticator.OAuth2
	authenticators = append(authenticators, setupOAuth2(s.configs.Auth.Google))
	//authenticators = append(authenticators, setupOAuth2(s.configs.Auth.Tweeter))
	s.oauth2Domain = domain.NewOAuth2Domain(
		s.userRepo,
		s.oauth2Repo,
		authenticators,
		s.configs.SessionSecret,
		s.configs.Auth.TokenSecret,
		s.configs.Auth.TokenExpiration,
	)
}

func (s *srv) loadControllers() {
	s.controllers = []controller{
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
	}

	for _, e := range s.controllers {
		e.Register(s.mux)
	}
}

func (s *srv) startServer() {
	fmt.Println("Starting server")
	s.mux.Handle("/", http.FileServer(http.Dir("./web")))
	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%s", s.configs.Server.Address, s.configs.Server.Port),
		Handler: s.mux,
	}

	if err := s.server.ListenAndServeTLS(s.configs.Server.Cert, s.configs.Server.Key); err != nil {
		panic(err)
	}
}

func setupOAuth2(config OAuth2Config) authenticator.OAuth2 {
	authenticator, err := authenticator.NewOAuth2(
		context.Background(),
		config.Name, config.Issuer,
		config.ClientID,
		config.ClientSecret,
	)
	if err != nil {
		panic(err)
	}

	return authenticator
}
