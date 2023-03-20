package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/internal/domains"
	"github.com/questx-lab/backend/internal/models"
	"github.com/questx-lab/backend/internal/repositories"

	_ "github.com/go-sql-driver/mysql"
)

type controller interface {
	Register(mux *http.ServeMux)
}

type srv struct {
	controllers []controller

	userRepo repositories.UserRepository

	userDomain domains.UserDomain
	authDomain domains.AuthDomain

	mux *http.ServeMux

	db *sql.DB

	configs *Configs

	server *http.Server
}

func (s *srv) loadMux() {
	s.mux = http.NewServeMux()
}

func (s *srv) loadConfig() {
	s.configs = &Configs{
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
	s.userRepo = repositories.NewUserRepository(s.db)
}

func (s *srv) loadDomains() {
	s.userDomain = domains.NewUserDomain(s.userRepo)
	s.authDomain = domains.NewAuthDomain(s.userRepo)
}

func (s *srv) loadControllers() {
	s.controllers = []controller{
		&api.Endpoint[models.LoginRequest, models.LoginResponse]{
			Path:   "/auth/login",
			Method: http.MethodPost,
			Handle: s.authDomain.Login,
		},
		&api.Endpoint[models.RegisterRequest, models.RegisterResponse]{
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
