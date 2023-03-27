package domain

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type CollaboratorDomain interface {
	Create(ctx router.Context, req *model.CreateCollaboratorRequest) (*model.CreateCollaboratorResponse, error)
	GetList(ctx router.Context, req *model.GetListCollaboratorRequest) (*model.GetListCollaboratorResponse, error)
	UpdateRole(ctx router.Context, req *model.UpdateCollaboratorRoleRequest) (*model.UpdateCollaboratorRoleResponse, error)
	Delete(ctx router.Context, req *model.DeleteCollaboratorRequest) (*model.DeleteCollaboratorResponse, error)
}

type collaboratorDomain struct {
	projectRepo      repository.ProjectRepository
	collaboratorRepo repository.CollaboratorRepository
	userRepo         repository.UserRepository
}

func NewCollaboratorDomain(
	projectRepo repository.ProjectRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
) CollaboratorDomain {
	return &collaboratorDomain{
		projectRepo:      projectRepo,
		collaboratorRepo: collaboratorRepo,
		userRepo:         userRepo,
	}
}

func (d *collaboratorDomain) Create(ctx router.Context, req *model.CreateCollaboratorRequest) (*model.CreateCollaboratorResponse, error) {
	userID := ctx.GetUserID()
	role := entity.Role(req.Role)

	//* users cannot assign by themselves
	if userID == req.UserID {
		return nil, errorx.NewGeneric(errorx.ErrPermissionDenied, "can not assign by yourself")
	}

	if !slices.Contains(entity.Roles, role) {
		return nil, errorx.NewGeneric(errorx.ErrBadRequest, "role is invalid")
	}

	// check user exists
	if _, err := d.userRepo.GetByID(ctx, req.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.NewGeneric(errorx.ErrNotFound, "user not found")
		}
		return nil, errorx.NewGeneric(errorx.ErrInternalServerError, err.Error())
	}

	// check project exists
	project, err := d.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.NewGeneric(errorx.ErrNotFound, "project not found")
		}
		return nil, errorx.NewGeneric(errorx.ErrInternalServerError, err.Error())
	}

	//! permission
	if err := verifyProjectPermission(ctx, d.collaboratorRepo, project.ID); err != nil {
		return nil, errorx.NewGeneric(errorx.ErrPermissionDenied, err.Error())
	}

	e := &entity.Collaborator{
		Base: entity.Base{
			ID: uuid.NewString(),
		},
		UserID:    req.UserID,
		CreatedBy: userID,
		Role:      role,
	}
	if err := d.collaboratorRepo.Create(ctx, e); err != nil {
		return nil, errorx.NewGeneric(errorx.ErrInternalServerError, err.Error())
	}
	return &model.CreateCollaboratorResponse{
		Success: true,
		ID:      e.ID,
	}, nil
}

func (d *collaboratorDomain) GetList(ctx router.Context, req *model.GetListCollaboratorRequest) (*model.GetListCollaboratorResponse, error) {
	entities, err := d.collaboratorRepo.GetList(ctx, req.Offset, req.Limit)
	if err != nil {
		return nil, errorx.NewGeneric(errorx.ErrInternalServerError, fmt.Errorf("unable to get list categories: %w", err).Error())
	}

	var data []*model.Collaborator
	for _, e := range entities {
		data = append(data, &model.Collaborator{
			ID:          e.ID,
			ProjectID:   e.Project.ID,
			UserID:      e.User.ID,
			UserName:    e.User.Name,
			ProjectName: e.Project.Name,
			Role:        string(e.Role),
		})
	}

	return &model.GetListCollaboratorResponse{
		Data:    data,
		Success: true,
	}, nil
}

func (d *collaboratorDomain) UpdateRole(ctx router.Context, req *model.UpdateCollaboratorRoleRequest) (*model.UpdateCollaboratorRoleResponse, error) {
	userID := ctx.GetUserID()
	role := entity.Role(req.Role)

	//* users cannot assign by themselves
	if userID == req.UserID {
		return nil, errorx.NewGeneric(errorx.ErrPermissionDenied, "can not assign by yourself")
	}

	if !slices.Contains(entity.Roles, role) {
		return nil, errorx.NewGeneric(errorx.ErrBadRequest, "role is invalid")
	}
	collaborator, err := d.collaboratorRepo.GetCollaborator(ctx, req.ProjectID, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.NewGeneric(errorx.ErrNotFound, "collaborator not found")
		}

		return nil, errorx.NewGeneric(errorx.ErrNotFound, fmt.Errorf("unable to retrieve collaborator: %w", err).Error())
	}

	if err := verifyProjectPermission(ctx, d.collaboratorRepo, collaborator.ProjectID); err != nil {
		return nil, errorx.NewGeneric(errorx.ErrPermissionDenied, err.Error())
	}

	if err := d.collaboratorRepo.UpdateRole(ctx, req.UserID, req.ProjectID, role); err != nil {
		return nil, fmt.Errorf("unable to update category: %w", err)
	}

	return &model.UpdateCollaboratorRoleResponse{
		Success: true,
	}, nil
}

func (d *collaboratorDomain) Delete(ctx router.Context, req *model.DeleteCollaboratorRequest) (*model.DeleteCollaboratorResponse, error) {
	userID := ctx.GetUserID()

	//* users cannot assign by themselves
	if userID == req.UserID {
		return nil, errorx.NewGeneric(errorx.ErrPermissionDenied, "can not assign by yourself")
	}

	collaborator, err := d.collaboratorRepo.GetCollaborator(ctx, req.ProjectID, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.NewGeneric(errorx.ErrNotFound, "collaborator not found")
		}

		return nil, errorx.NewGeneric(errorx.ErrNotFound, fmt.Errorf("unable to retrieve collaborator: %w", err).Error())
	}

	if err := verifyProjectPermission(ctx, d.collaboratorRepo, collaborator.ProjectID); err != nil {
		return nil, errorx.NewGeneric(errorx.ErrPermissionDenied, err.Error())
	}

	if err := d.collaboratorRepo.Delete(ctx, req.UserID, req.ProjectID); err != nil {
		return nil, errorx.NewGeneric(errorx.ErrInternalServerError, fmt.Errorf("unable to update category: %w", err).Error())
	}

	return &model.DeleteCollaboratorResponse{
		Success: true,
	}, nil
}
