package domain

import (
	"time"

	"github.com/questx-lab/backend/api"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"

	"github.com/google/uuid"
)

type ProjectDomain interface {
	Create(ctx *api.Context, req *model.CreateProjectRequest) (*model.CreateProjectResponse, error)
	GetList(ctx *api.Context, req *model.GetListProjectRequest) (*model.GetListProjectResponse, error)
	GeyByID(ctx *api.Context, req *model.GetProjectByIDRequest) (*model.GetProjectByIDResponse, error)
	UpdateByID(ctx *api.Context, req *model.UpdateProjectByIDRequest) (*model.UpdateProjectByIDResponse, error)
	DeleteByID(ctx *api.Context, req *model.DeleteProjectByIDRequest) (*model.DeleteProjectByIDResponse, error)
}

type projectDomain struct {
	projectRepo repository.ProjectRepository
}

func NewProjectDomain(projectRepo repository.ProjectRepository) ProjectDomain {
	return &projectDomain{
		projectRepo: projectRepo,
	}
}

func (d *projectDomain) Create(ctx *api.Context, req *model.CreateProjectRequest) (*model.CreateProjectResponse, error) {
	now := time.Now()
	userID, err := ctx.ExtractUserIDFromContext()
	if err != nil {
		return nil, err
	}
	if err := d.projectRepo.Create(ctx.Context, &entity.Project{
		Base: entity.Base{
			ID:        uuid.NewString(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Twitter:   req.Twitter,
		Discord:   req.Discord,
		Telegram:  req.Telegram,
		CreatedBy: userID,
	}); err != nil {
		return nil, err
	}

	return &model.CreateProjectResponse{
		Response: model.Response{
			Code:    200,
			Success: true,
		},
	}, nil
}

func (d *projectDomain) GetList(ctx *api.Context, req *model.GetListProjectRequest) (*model.GetListProjectResponse, error) {
	result, err := d.projectRepo.GetList(ctx.Context, req.Offset, req.Limit)
	if err != nil {
		return nil, err
	}

	return &model.GetListProjectResponse{
		Response: model.Response{
			Code:    200,
			Success: true,
		},
		Data: result,
	}, nil
}

func (d *projectDomain) GeyByID(ctx *api.Context, req *model.GetProjectByIDRequest) (*model.GetProjectByIDResponse, error) {
	result, err := d.projectRepo.GeyByID(ctx.Context, req.ID)
	if err != nil {
		return nil, err
	}

	return &model.GetProjectByIDResponse{
		Response: model.Response{
			Code:    200,
			Success: true,
		},
		Data: result,
	}, nil
}

func (d *projectDomain) UpdateByID(ctx *api.Context, req *model.UpdateProjectByIDRequest) (*model.UpdateProjectByIDResponse, error) {
	err := d.projectRepo.UpdateByID(ctx.Context, req.ID, &entity.Project{
		Twitter:  req.Twitter,
		Telegram: req.Telegram,
		Discord:  req.Discord,
		Base: entity.Base{
			UpdatedAt: time.Now(),
		},
	})
	if err != nil {
		return nil, err
	}

	return &model.UpdateProjectByIDResponse{
		Response: model.Response{
			Code:    200,
			Success: true,
		},
	}, nil
}

func (d *projectDomain) DeleteByID(ctx *api.Context, req *model.DeleteProjectByIDRequest) (*model.DeleteProjectByIDResponse, error) {
	if err := d.projectRepo.DeleteByID(ctx.Context, req.ID); err != nil {
		return nil, err
	}

	return &model.DeleteProjectByIDResponse{
		Response: model.Response{
			Code:    200,
			Success: true,
		},
	}, nil
}
