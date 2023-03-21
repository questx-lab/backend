package domain

import (
	"time"

	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/multierr"
)

type ProjectDomain interface {
	CreateProject(api.CustomContext, *model.CreateProjectRequest) (*model.CreateProjectResponse, error)
}

type projectDomain struct {
	projectRepo repository.ProjectRepository
}

func NewProjectDomain(projectRepo repository.ProjectRepository) ProjectDomain {
	return &projectDomain{projectRepo: projectRepo}
}

func (d *projectDomain) CreateProject(ctx api.CustomContext, req *model.CreateProjectRequest) (*model.CreateProjectResponse, error) {
	now := time.Now()
	e := &entity.Project{}

	userID := ctx.ExtractUserIDFromContext()

	if err := multierr.Combine(
		e.ID.Scan(uuid.NewString()),
		e.CreatedAt.Scan(now),
		e.UpdatedAt.Scan(now),
		e.Twitter.Scan(req.Twitter),
		e.Discord.Scan(req.Discord),
		e.Telegram.Scan(req.Telegram),
		e.Name.Scan(req.Name),
		e.CreatedBy.Scan(userID),
		e.DeletedAt.Scan(nil),
	); err != nil {
		return nil, err
	}
	if err := d.projectRepo.Create(ctx, e); err != nil {
		return nil, err
	}
	return &model.CreateProjectResponse{}, nil
}
