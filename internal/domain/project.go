package domain

import (
	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
)

type ProjectDomain interface {
	CreateProject(api.CustomContext, *model.CreateProjectRequest) (*model.CreateProjectResponse, error)
}

type projectDomain struct {
	projectRepo repository.UserRepository
}
