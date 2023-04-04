package domain

import (
	"errors"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type CollaboratorDomain interface {
	Create(ctx xcontext.Context, req *model.CreateCollaboratorRequest) (*model.CreateCollaboratorResponse, error)
	GetList(ctx xcontext.Context, req *model.GetListCollaboratorRequest) (*model.GetListCollaboratorResponse, error)
	UpdateRole(ctx xcontext.Context, req *model.UpdateCollaboratorRoleRequest) (*model.UpdateCollaboratorRoleResponse, error)
	Delete(ctx xcontext.Context, req *model.DeleteCollaboratorRequest) (*model.DeleteCollaboratorResponse, error)
}

type collaboratorDomain struct {
	projectRepo      repository.ProjectRepository
	collaboratorRepo repository.CollaboratorRepository
	userRepo         repository.UserRepository
	roleVerifier     *projectRoleVerifier
}

func NewCollaboratorDomain(
	projectRepo repository.ProjectRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
) CollaboratorDomain {
	return &collaboratorDomain{
		projectRepo:      projectRepo,
		userRepo:         userRepo,
		collaboratorRepo: collaboratorRepo,
		roleVerifier:     newProjectRoleVerifier(collaboratorRepo),
	}
}

func (d *collaboratorDomain) Create(ctx xcontext.Context, req *model.CreateCollaboratorRequest) (*model.CreateCollaboratorResponse, error) {
	userID := xcontext.GetRequestUserID(ctx)

	// users cannot assign by themselves
	if userID == req.UserID {
		return nil, errorx.New(errorx.PermissionDenied, "Can not assign by yourself")
	}

	role, err := enum.ToEnum[entity.Role](req.Role)
	if err != nil {
		ctx.Logger().Debugf("Invalid role %s: %v", req.Role, err)
		return nil, errorx.New(errorx.BadRequest, "Invalid role")
	}

	// check user exists
	if _, err := d.userRepo.GetByID(ctx, req.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found user")
		}

		ctx.Logger().Errorf("Cannot get user: %v", err)
		return nil, errorx.Unknown
	}

	// check project exists
	project, err := d.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found project")
		}

		ctx.Logger().Errorf("Cannot get project: %v", err)
		return nil, errorx.Unknown
	}

	// permission
	if err = d.roleVerifier.Verify(ctx, project.ID, entity.AdminGroup...); err != nil {
		ctx.Logger().Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
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
		ctx.Logger().Errorf("Cannot create collaborator: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateCollaboratorResponse{ID: e.ID}, nil
}

func (d *collaboratorDomain) GetList(ctx xcontext.Context, req *model.GetListCollaboratorRequest) (*model.GetListCollaboratorResponse, error) {
	entities, err := d.collaboratorRepo.GetList(ctx, req.Offset, req.Limit)
	if err != nil {
		ctx.Logger().Errorf("Cannot get list of collaborator: %v", err)
		return nil, errorx.Unknown
	}

	var data []model.Collaborator
	for _, e := range entities {
		data = append(data, model.Collaborator{
			ID:          e.ID,
			ProjectID:   e.Project.ID,
			UserID:      e.User.ID,
			UserName:    e.User.Name,
			ProjectName: e.Project.Name,
			Role:        string(e.Role),
		})
	}

	return &model.GetListCollaboratorResponse{Collaborators: data}, nil
}

func (d *collaboratorDomain) UpdateRole(ctx xcontext.Context, req *model.UpdateCollaboratorRoleRequest) (*model.UpdateCollaboratorRoleResponse, error) {
	userID := xcontext.GetRequestUserID(ctx)

	// users cannot assign by themselves
	if userID == req.UserID {
		return nil, errorx.New(errorx.PermissionDenied, "Can not assign by yourself")
	}

	role, err := enum.ToEnum[entity.Role](req.Role)
	if err != nil {
		ctx.Logger().Debugf("Invalid role %s: %v", req.Role, err)
		return nil, errorx.New(errorx.BadRequest, "Invalid role")
	}

	collaborator, err := d.collaboratorRepo.Get(ctx, req.ProjectID, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found collaborator")
		}

		ctx.Logger().Errorf("Cannot get collaborator: %v", err)
		return nil, errorx.Unknown
	}

	if err = d.roleVerifier.Verify(ctx, collaborator.ProjectID, entity.AdminGroup...); err != nil {
		ctx.Logger().Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if err := d.collaboratorRepo.UpdateRole(ctx, req.UserID, req.ProjectID, role); err != nil {
		ctx.Logger().Errorf("Cannot update collaborator role: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateCollaboratorRoleResponse{}, nil
}

func (d *collaboratorDomain) Delete(ctx xcontext.Context, req *model.DeleteCollaboratorRequest) (*model.DeleteCollaboratorResponse, error) {
	userID := xcontext.GetRequestUserID(ctx)

	// users cannot assign by themselves
	if userID == req.UserID {
		return nil, errorx.New(errorx.PermissionDenied, "Can not assign by yourself")
	}

	collaborator, err := d.collaboratorRepo.Get(ctx, req.ProjectID, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found collaborator")
		}

		ctx.Logger().Errorf("Cannot get collaborator: %v", err)
		return nil, errorx.Unknown
	}

	if err = d.roleVerifier.Verify(ctx, collaborator.ProjectID, entity.AdminGroup...); err != nil {
		ctx.Logger().Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if err := d.collaboratorRepo.Delete(ctx, req.UserID, req.ProjectID); err != nil {
		ctx.Logger().Errorf("Cannot delete collaborator: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteCollaboratorResponse{}, nil
}
