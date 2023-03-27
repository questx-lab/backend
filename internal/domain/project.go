package domain

import (
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"

	"github.com/google/uuid"
)

type ProjectDomain interface {
	Create(ctx router.Context, req *model.CreateProjectRequest) (*model.CreateProjectResponse, error)
	GetList(ctx router.Context, req *model.GetListProjectRequest) (*model.GetListProjectResponse, error)
	GetByID(ctx router.Context, req *model.GetProjectByIDRequest) (*model.GetProjectByIDResponse, error)
	UpdateByID(ctx router.Context, req *model.UpdateProjectByIDRequest) (*model.UpdateProjectByIDResponse, error)
	DeleteByID(ctx router.Context, req *model.DeleteProjectByIDRequest) (*model.DeleteProjectByIDResponse, error)
}

type projectDomain struct {
	projectRepo      repository.ProjectRepository
	collaboratorRepo repository.CollaboratorRepository
}

func NewProjectDomain(projectRepo repository.ProjectRepository, collaboratorRepo repository.CollaboratorRepository) ProjectDomain {
	return &projectDomain{
		projectRepo:      projectRepo,
		collaboratorRepo: collaboratorRepo,
	}
}

func (d *projectDomain) Create(ctx router.Context, req *model.CreateProjectRequest) (
	*model.CreateProjectResponse, error) {
	userID := ctx.GetUserID()
	proj := &entity.Project{
		Base: entity.Base{
			ID: uuid.NewString(),
		},
		Name:      req.Name,
		Twitter:   req.Twitter,
		Discord:   req.Discord,
		Telegram:  req.Telegram,
		CreatedBy: userID,
	}
	if err := d.projectRepo.Create(ctx, proj); err != nil {
		return nil, errorx.NewGeneric(err, "cannot create project")
	}

	if err := d.collaboratorRepo.Create(ctx, &entity.Collaborator{
		Base: entity.Base{
			ID: uuid.NewString(),
		},
		UserID:    userID,
		ProjectID: proj.ID,
		Role:      entity.Owner,
	}); err != nil {
		return nil, errorx.NewGeneric(err, "cannot create project")
	}

	return &model.CreateProjectResponse{
		Response: model.Response{
			Success: true,
		},
		ID: proj.ID,
	}, nil
}

func (d *projectDomain) GetList(ctx router.Context, req *model.GetListProjectRequest) (
	*model.GetListProjectResponse, error) {
	result, err := d.projectRepo.GetList(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, errorx.NewGeneric(err, "cannot get project list")
	}

	return &model.GetListProjectResponse{
		Response: model.Response{
			Success: true,
		},
		Data: result,
	}, nil
}

func (d *projectDomain) GetByID(ctx router.Context, req *model.GetProjectByIDRequest) (
	*model.GetProjectByIDResponse, error) {
	result, err := d.projectRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, errorx.NewGeneric(err, "cannot get project")
	}

	return &model.GetProjectByIDResponse{
		Response: model.Response{
			Success: true,
		},
		Data: result,
	}, nil
}

func (d *projectDomain) UpdateByID(ctx router.Context, req *model.UpdateProjectByIDRequest) (
	*model.UpdateProjectByIDResponse, error) {
	err := d.projectRepo.UpdateByID(ctx, req.ID, &entity.Project{
		Twitter:  req.Twitter,
		Telegram: req.Telegram,
		Discord:  req.Discord,
		Base: entity.Base{
			UpdatedAt: time.Now(),
		},
	})
	if err != nil {
		return nil, errorx.NewGeneric(err, "cannot update project")
	}

	return &model.UpdateProjectByIDResponse{
		Response: model.Response{
			Success: true,
		},
	}, nil
}

func (d *projectDomain) DeleteByID(ctx router.Context, req *model.DeleteProjectByIDRequest) (
	*model.DeleteProjectByIDResponse, error) {
	if err := d.projectRepo.DeleteByID(ctx, req.ID); err != nil {
		return nil, errorx.NewGeneric(err, "cannot delete project")
	}

	return &model.DeleteProjectByIDResponse{
		Response: model.Response{
			Success: true,
		},
	}, nil
}
