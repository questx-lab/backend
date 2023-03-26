package domain

import (
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/router"
)

type CollaboratorDomain interface {
	Create(ctx router.Context, req *model.CreateCollaboratorRequest) (*model.CreateCollaboratorResponse, error)
	GetList(ctx router.Context, req *model.GetListCollaboratorRequest) (*model.GetListCollaboratorResponse, error)
	UpdateByID(ctx router.Context, req *model.UpdateCollaboratorByIDRequest) (*model.UpdateCollaboratorByIDResponse, error)
	DeleteByID(ctx router.Context, req *model.DeleteCollaboratorByIDRequest) (*model.DeleteCollaboratorByIDResponse, error)
}

type collaboratorDomain struct {
	projectRepo      repository.ProjectRepository
	collaboratorRepo repository.CollaboratorRepository
}

func NewCollaboratorDomain(
	projectRepo repository.ProjectRepository,
	collaboratorRepo repository.CollaboratorRepository,
) CollaboratorDomain {
	return &collaboratorDomain{
		projectRepo:      projectRepo,
		collaboratorRepo: collaboratorRepo,
	}
}

func (d *collaboratorDomain) Create(ctx router.Context, req *model.CreateCollaboratorRequest) (*model.CreateCollaboratorResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (d *collaboratorDomain) GetList(ctx router.Context, req *model.GetListCollaboratorRequest) (*model.GetListCollaboratorResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (d *collaboratorDomain) UpdateByID(ctx router.Context, req *model.UpdateCollaboratorByIDRequest) (*model.UpdateCollaboratorByIDResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (d *collaboratorDomain) DeleteByID(ctx router.Context, req *model.DeleteCollaboratorByIDRequest) (*model.DeleteCollaboratorByIDResponse, error) {
	panic("not implemented") // TODO: Implement
}
