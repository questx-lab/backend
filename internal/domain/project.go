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
		ctx.Logger().Errorf("Cannot create project: %v", err)
		return nil, errorx.Unknown
	}

	err := d.collaboratorRepo.Create(ctx, &entity.Collaborator{
		Base:      entity.Base{ID: uuid.NewString()},
		UserID:    userID,
		ProjectID: proj.ID,
		Role:      entity.Owner,
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot assign role owner: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateProjectResponse{ID: proj.ID}, nil
}

func (d *projectDomain) GetList(ctx router.Context, req *model.GetListProjectRequest) (
	*model.GetListProjectResponse, error) {
	result, err := d.projectRepo.GetList(ctx, req.Offset, req.Limit)
	if err != nil {
		ctx.Logger().Errorf("Cannot get project list: %v", err)
		return nil, errorx.Unknown
	}

	projects := []model.Project{}
	for _, p := range result {
		projects = append(projects, model.Project{
			ID:        p.ID,
			CreatedAt: p.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt: p.UpdatedAt.Format(time.RFC3339Nano),
			CreatedBy: p.CreatedBy,
			Name:      p.Name,
			Twitter:   p.Twitter,
			Telegram:  p.Telegram,
			Discord:   p.Discord,
		})
	}

	return &model.GetListProjectResponse{Projects: projects}, nil
}

func (d *projectDomain) GetByID(ctx router.Context, req *model.GetProjectByIDRequest) (
	*model.GetProjectByIDResponse, error) {
	result, err := d.projectRepo.GetByID(ctx, req.ID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get the project: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetProjectByIDResponse{Project: model.Project{
		ID:        result.ID,
		CreatedAt: result.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: result.UpdatedAt.Format(time.RFC3339Nano),
		CreatedBy: result.CreatedBy,
		Name:      result.Name,
		Twitter:   result.Twitter,
		Telegram:  result.Telegram,
		Discord:   result.Discord,
	}}, nil
}

func (d *projectDomain) UpdateByID(ctx router.Context, req *model.UpdateProjectByIDRequest) (
	*model.UpdateProjectByIDResponse, error) {
	err := d.projectRepo.UpdateByID(ctx, req.ID, &entity.Project{
		Twitter:  req.Twitter,
		Telegram: req.Telegram,
		Discord:  req.Discord,
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot update project: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateProjectByIDResponse{}, nil
}

func (d *projectDomain) DeleteByID(ctx router.Context, req *model.DeleteProjectByIDRequest) (
	*model.DeleteProjectByIDResponse, error) {
	if err := d.projectRepo.DeleteByID(ctx, req.ID); err != nil {
		ctx.Logger().Errorf("Cannot delete project: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteProjectByIDResponse{}, nil
}
