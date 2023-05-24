package domain

import (
	"context"
	"errors"
	"time"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type CollaboratorDomain interface {
	Assign(ctx context.Context, req *model.AssignCollaboratorRequest) (*model.AssignCollaboratorResponse, error)
	GetMyCollabs(context.Context, *model.GetMyCollabsRequest) (*model.GetMyCollabsResponse, error)
	GetCommunityCollabs(ctx context.Context, req *model.GetCommunityCollabsRequest) (*model.GetCommunityCollabsResponse, error)
	Delete(ctx context.Context, req *model.DeleteCollaboratorRequest) (*model.DeleteCollaboratorResponse, error)
}

type collaboratorDomain struct {
	communityRepo    repository.CommunityRepository
	collaboratorRepo repository.CollaboratorRepository
	userRepo         repository.UserRepository
	roleVerifier     *common.CommunityRoleVerifier
}

func NewCollaboratorDomain(
	communityRepo repository.CommunityRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
) CollaboratorDomain {
	return &collaboratorDomain{
		communityRepo:    communityRepo,
		userRepo:         userRepo,
		collaboratorRepo: collaboratorRepo,
		roleVerifier:     common.NewCommunityRoleVerifier(collaboratorRepo, userRepo),
	}
}

func (d *collaboratorDomain) Assign(
	ctx context.Context, req *model.AssignCollaboratorRequest,
) (*model.AssignCollaboratorResponse, error) {
	// user cannot assign by themselves
	if xcontext.RequestUserID(ctx) == req.UserID {
		return nil, errorx.New(errorx.PermissionDenied, "Can not assign by yourself")
	}

	role, err := enum.ToEnum[entity.Role](req.Role)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid role %s: %v", req.Role, err)
		return nil, errorx.New(errorx.BadRequest, "Invalid role")
	}

	var needRole []entity.Role
	switch role {
	case entity.Owner:
		return nil, errorx.New(errorx.PermissionDenied, "Cannot set the owner")
	case entity.Editor:
		needRole = []entity.Role{entity.Owner}
	case entity.Reviewer:
		needRole = entity.AdminGroup
	default:
		return nil, errorx.New(errorx.BadRequest, "Invalid role %s", role)
	}

	// Check permission of the user giving the role against to that role.
	if err = d.roleVerifier.Verify(ctx, req.CommunityID, needRole...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	currentCollab, err := d.collaboratorRepo.Get(ctx, req.CommunityID, req.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		xcontext.Logger(ctx).Errorf("Cannot get current collaborator of user: %v", err)
		return nil, errorx.Unknown
	}

	if err == nil {
		needRole = nil
		switch currentCollab.Role {
		case entity.Owner:
			return nil, errorx.New(errorx.PermissionDenied, "No one can assign another role to owner")
		case entity.Editor:
			needRole = []entity.Role{entity.Owner}
		case entity.Reviewer:
			needRole = entity.AdminGroup
		}

		// Check permission of the user giving role against to the user
		// receiving the role.
		if len(needRole) > 0 {
			if err = d.roleVerifier.Verify(ctx, req.CommunityID, needRole...); err != nil {
				xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
				return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
			}
		}
	}

	e := &entity.Collaborator{
		UserID:      req.UserID,
		CommunityID: req.CommunityID,
		Role:        role,
		CreatedBy:   xcontext.RequestUserID(ctx),
	}
	if err := d.collaboratorRepo.Upsert(ctx, e); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create collaborator: %v", err)
		return nil, errorx.Unknown
	}

	return &model.AssignCollaboratorResponse{}, nil
}

func (d *collaboratorDomain) Delete(
	ctx context.Context, req *model.DeleteCollaboratorRequest,
) (*model.DeleteCollaboratorResponse, error) {
	// user cannot delete by themselves
	if xcontext.RequestUserID(ctx) == req.UserID {
		return nil, errorx.New(errorx.PermissionDenied, "Can not delete by yourself")
	}

	collaborator, err := d.collaboratorRepo.Get(ctx, req.CommunityID, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found collaborator")
		}

		xcontext.Logger(ctx).Errorf("Cannot get collaborator: %v", err)
		return nil, errorx.Unknown
	}

	var needRole []entity.Role
	switch collaborator.Role {
	case entity.Owner:
		return nil, errorx.New(errorx.PermissionDenied, "Cannot delete the owner")
	case entity.Editor:
		needRole = []entity.Role{entity.Owner}
	case entity.Reviewer:
		needRole = entity.AdminGroup
	default:
		xcontext.Logger(ctx).Errorf("Invalid role in database: %s", collaborator.Role)
		return nil, errorx.Unknown
	}

	if err = d.roleVerifier.Verify(ctx, collaborator.CommunityID, needRole...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if err := d.collaboratorRepo.Delete(ctx, req.UserID, req.CommunityID); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot delete collaborator: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteCollaboratorResponse{}, nil
}

func (d *collaboratorDomain) GetCommunityCollabs(
	ctx context.Context, req *model.GetCommunityCollabsRequest,
) (*model.GetCommunityCollabsResponse, error) {
	// Any collaborator of community can see other ones.
	_, err := d.collaboratorRepo.Get(ctx, req.Community, xcontext.RequestUserID(ctx))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
		}
		xcontext.Logger(ctx).Errorf("Cannot get collaborator: %v", err)
		return nil, errorx.Unknown
	}

	entities, err := d.collaboratorRepo.GetListByCommunityID(ctx, req.Community, req.Offset, req.Limit)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get list of collaborator: %v", err)
		return nil, errorx.Unknown
	}

	data := []model.Collaborator{}
	for _, e := range entities {
		data = append(data, model.Collaborator{
			CommunityID: e.Community.ID,
			UserID:      e.UserID,
			User: model.User{
				ID:        e.User.ID,
				Name:      e.User.Name,
				Address:   e.User.Address.String,
				Role:      string(e.User.Role),
				AvatarURL: e.User.ProfilePicture,
			},
			Role:      string(e.Role),
			CreatedBy: e.CreatedBy,
		})
	}

	return &model.GetCommunityCollabsResponse{Collaborators: data}, nil
}

func (d *collaboratorDomain) GetMyCollabs(
	ctx context.Context, req *model.GetMyCollabsRequest,
) (*model.GetMyCollabsResponse, error) {
	if req.Limit == 0 {
		req.Limit = -1
	}

	userID := xcontext.RequestUserID(ctx)
	result, err := d.collaboratorRepo.GetListByUserID(ctx, userID, req.Offset, req.Limit)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get community list: %v", err)
		return nil, errorx.Unknown
	}

	collaborators := []model.Collaborator{}
	for _, collab := range result {
		collaborators = append(collaborators, model.Collaborator{
			CommunityID: collab.CommunityID,
			Community: model.Community{
				ID:           collab.Community.ID,
				CreatedAt:    collab.Community.CreatedAt.Format(time.RFC3339Nano),
				UpdatedAt:    collab.Community.UpdatedAt.Format(time.RFC3339Nano),
				CreatedBy:    collab.Community.CreatedBy,
				Introduction: string(collab.Community.Introduction),
				Name:         collab.Community.Name,
				Twitter:      collab.Community.Twitter,
				Discord:      collab.Community.Discord,
			},
			UserID:    userID,
			Role:      string(collab.Role),
			CreatedBy: collab.CreatedBy,
		})
	}

	return &model.GetMyCollabsResponse{Collaborators: collaborators}, nil
}
