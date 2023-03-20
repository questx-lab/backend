package main

import (
	"net/http"

	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/internal/domains"
	"github.com/questx-lab/backend/internal/models"
)

type register interface {
	Register(mux *http.ServeMux)
}

type srv struct {
	endpoints  []register
	userDomain domains.UserDomain
	authDomain domains.AuthDomain

	mux *http.ServeMux
}

func (s *srv) loadMux() {
	s.mux = http.NewServeMux()
}

func (s *srv) loadHandler() {
	s.endpoints = []register{
		&api.Endpoint[models.LoginRequest, models.LoginResponse]{
			Path:   "/auth/login",
			Method: http.MethodPost,
			Handle: s.authDomain.Login,
		},
		&api.Endpoint[models.RegisterRequest, models.RegisterResponse]{
			Path:   "/auth/login",
			Method: http.MethodPost,
			Handle: s.authDomain.Register,
		},
	}

	for _, e := range s.endpoints {
		e.Register(s.mux)
	}
}
