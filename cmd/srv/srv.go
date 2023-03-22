package main

import (
	"fmt"
	"log"
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
	authGroup := &api.Group{
		Path:   "/auth",
		Before: []api.Handler{api.Logger},
		After:  []api.Handler{api.Close},
	}

	projectGroup := &api.Group{
		Path: "/projects",
		Before: []api.Handler{
			api.Logger,
			api.ImportUserIDToContext(s.tknGenerator),
		},
		After: []api.Handler{api.Close},
	}

	s.controllers = []controller{
		&api.Endpoint[model.LoginRequest, model.LoginResponse]{
			Group:  authGroup,
			Path:   "/login",
			Method: http.MethodPost,
			Handle: s.authDomain.Login,
		},

		&api.Endpoint[model.RegisterRequest, model.RegisterResponse]{
			Group:  projectGroup,
			Path:   "/register",
			Method: http.MethodPost,
			Handle: s.authDomain.Register,
		},

		&api.Endpoint[model.CreateProjectRequest, model.CreateProjectResponse]{
			Group:  projectGroup,
			Path:   "/",
			Method: http.MethodPost,
			Handle: s.projectDomain.CreateProject,
			Before: []api.Handler{},
		},
	}
	for _, c := range s.controllers {
		c.Register(s.mux)
	}
}

func (s *srv) startServer() {
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.configs.Port),
		Handler: s.mux,
	}
	log.Printf("Starting server on port: %s\n", s.configs.Port)
	if err := s.server.ListenAndServe(); err != nil {
		panic(err)
	}
}
