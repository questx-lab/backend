package main

import (
	"database/sql"
	"net/http"

	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/internal/domains"
	"github.com/questx-lab/backend/internal/models"
	"github.com/questx-lab/backend/internal/repositories"
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
}

func (s *srv) loadMux() {
	s.mux = http.NewServeMux()
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
