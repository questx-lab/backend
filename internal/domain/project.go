package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
)

type ProjectDomain interface {
	CreateProject(*api.Context, *model.CreateProjectRequest) (*model.CreateProjectResponse, error)
}

type projectDomain struct {
	projectRepo repository.ProjectRepository
}

func NewProjectDomain(projectRepo repository.ProjectRepository) ProjectDomain {
	return &projectDomain{projectRepo: projectRepo}
}

func (d *projectDomain) CreateProject(ctx *api.Context, req *model.CreateProjectRequest) (*model.CreateProjectResponse, error) {
	now := time.Now()

	userID := ctx.ExtractUserIDFromContext()
	e := &entity.Project{
		ID:        uuid.NewString(),
		CreatedAt: now,
		UpdatedAt: now,
		Twitter:   req.Twitter,
		Telegram:  req.Telegram,
		Discord:   req.Discord,
		CreatedBy: userID,
	}

	if err := d.projectRepo.Create(ctx, e); err != nil {
		return nil, err
	}
	return &model.CreateProjectResponse{}, nil
}
