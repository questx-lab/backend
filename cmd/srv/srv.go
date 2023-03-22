package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/domain"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/utils/token"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type controller interface {
	Register(mux *http.ServeMux)
}

type srv struct {
	tknGenerator token.Generator

	controllers []controller

	userRepo    repository.UserRepository
	projectRepo repository.ProjectRepository

	userDomain    domain.UserDomain
	authDomain    domain.AuthDomain
	projectDomain domain.ProjectDomain

	mux *http.ServeMux

	db *gorm.DB

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
	s.db, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       s.configs.DBConnection, // data source name
		DefaultStringSize:         256,                    // default size for string fields
		DisableDatetimePrecision:  true,                   // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,                   // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,                   // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false,                  // auto configure based on currently MySQL version
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
}

func (s *srv) loadRepos() {
	// s.userRepo = repository.NewUserRepository(s.db)
	s.projectRepo = repository.NewProjectRepository(s.db)
}

func (s *srv) loadDomains() {
	s.userDomain = domain.NewUserDomain(s.userRepo)
	s.authDomain = domain.NewAuthDomain(s.userRepo)
	s.projectDomain = domain.NewProjectDomain(s.projectRepo)
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

		&api.Endpoint[model.CreateProjectRequest, model.CreateProjectResponse]{
			Path:   "/projects",
			Method: http.MethodPost,
			Handle: s.projectDomain.CreateProject,
			Before: []api.Handler{
				api.ImportUserIDToContext(s.tknGenerator),
			},
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
