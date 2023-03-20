package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/domain"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"

	_ "github.com/go-sql-driver/mysql"
)

type controller interface {
	Register(mux *http.ServeMux)
}

type srv struct {
	controllers []controller

	userRepo repository.UserRepository

	userDomain domain.UserDomain
	authDomain domain.AuthDomain

	mux *http.ServeMux

	db *sql.DB

	configs *config.Configs

	server *http.Server
}

func (s *srv) loadMux() {
	s.mux = http.NewServeMux()
}

func (s *srv) loadConfig() {
	s.configs = &config.Configs{
		DBConnection: os.Getenv("DB_CONNECTION"),
		Port:         os.Getenv("PORT"),
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
}

func (s *srv) loadDomains() {
	s.userDomain = domain.NewUserDomain(s.userRepo)
	s.authDomain = domain.NewAuthDomain(s.userRepo)
}

func (s *srv) loadControllers() {
	s.controllers = []controller{
		&api.Endpoint[model.LoginRequest, model.LoginResponse]{
			Path:   "/auth/login",
			Method: http.MethodPost,
			Handle: s.authDomain.Login,
		},
		&api.Endpoint[model.RegisterRequest, model.RegisterResponse]{
			Path:   "/auth/register",
			Method: http.MethodPost,
			Handle: s.authDomain.Register,
		},
	}

	for _, e := range s.controllers {
		e.Register(s.mux)
	}
}

func (s *srv) startServer() {
	fmt.Println("Starting server")
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.configs.Port),
		Handler: s.mux,
	}
	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
}
