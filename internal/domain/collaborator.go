package domain

import (
	"context"
	"errors"

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
	questRepo        repository.QuestRepository
	roleVerifier     *common.CommunityRoleVerifier
}

func NewCollaboratorDomain(
	communityRepo repository.CommunityRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
	questRepo repository.QuestRepository,
) CollaboratorDomain {
	return &collaboratorDomain{
		communityRepo:    communityRepo,
		userRepo:         userRepo,
		collaboratorRepo: collaboratorRepo,
		questRepo:        questRepo,
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

	role, err := enum.ToEnum[entity.CollaboratorRole](req.Role)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid role %s: %v", req.Role, err)
		return nil, errorx.New(errorx.BadRequest, "Invalid role")
	}

	var needRole []entity.CollaboratorRole
	switch role {
	case entity.Owner:
		return nil, errorx.New(errorx.PermissionDenied, "Cannot set the owner")
	case entity.Editor:
		needRole = []entity.CollaboratorRole{entity.Owner}
	case entity.Reviewer:
		needRole = entity.AdminGroup
	default:
		return nil, errorx.New(errorx.BadRequest, "Invalid role %s", role)
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	// Check permission of the user giving the role against to that role.
	if err = d.roleVerifier.Verify(ctx, community.ID, needRole...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	currentCollab, err := d.collaboratorRepo.Get(ctx, community.ID, req.UserID)
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
			needRole = []entity.CollaboratorRole{entity.Owner}
		case entity.Reviewer:
			needRole = entity.AdminGroup
		}

		// Check permission of the user giving role against to the user
		// receiving the role.
		if len(needRole) > 0 {
			if err = d.roleVerifier.Verify(ctx, community.ID, needRole...); err != nil {
				xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
				return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
			}
		}
	}

	e := &entity.Collaborator{
		UserID:      req.UserID,
		CommunityID: community.ID,
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

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	collaborator, err := d.collaboratorRepo.Get(ctx, community.ID, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found collaborator")
		}

		xcontext.Logger(ctx).Errorf("Cannot get collaborator: %v", err)
		return nil, errorx.Unknown
	}

	var needRole []entity.CollaboratorRole
	switch collaborator.Role {
	case entity.Owner:
		return nil, errorx.New(errorx.PermissionDenied, "Cannot delete the owner")
	case entity.Editor:
		needRole = []entity.CollaboratorRole{entity.Owner}
	case entity.Reviewer:
		needRole = entity.AdminGroup
	default:
		xcontext.Logger(ctx).Errorf("Invalid role in database: %s", collaborator.Role)
		return nil, errorx.Unknown
	}

	if err = d.roleVerifier.Verify(ctx, community.ID, needRole...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	if err := d.collaboratorRepo.Delete(ctx, req.UserID, community.ID); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot delete collaborator: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteCollaboratorResponse{}, nil
}

func (d *collaboratorDomain) GetCommunityCollabs(
	ctx context.Context, req *model.GetCommunityCollabsRequest,
) (*model.GetCommunityCollabsResponse, error) {
	// The number records of this API is small. No need to check limit.
	if req.Limit == 0 {
		req.Limit = -1
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	// Any collaborator of community can see other ones.
	if _, err := d.collaboratorRepo.Get(ctx, community.ID, xcontext.RequestUserID(ctx)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
		}
		xcontext.Logger(ctx).Errorf("Cannot get collaborator: %v", err)
		return nil, errorx.Unknown
	}

	collaborators, err := d.collaboratorRepo.GetListByCommunityID(ctx, community.ID, req.Offset, req.Limit)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get list of collaborator: %v", err)
		return nil, errorx.Unknown
	}

	data := []model.Collaborator{}
	for _, c := range collaborators {
		data = append(data,
			convertCollaborator(
				&c,
				model.Community{Handle: community.Handle},
				convertUser(&c.User, nil, false),
			),
		)
	}

	return &model.GetCommunityCollabsResponse{Collaborators: data}, nil
}

func (d *collaboratorDomain) GetMyCollabs(
	ctx context.Context, req *model.GetMyCollabsRequest,
) (*model.GetMyCollabsResponse, error) {
	// The number records of this API is small. No need to check limit.
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
		totalQuests, err := d.questRepo.Count(
			ctx, repository.StatisticQuestFilter{CommunityID: collab.Community.ID})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot count quest of community %s: %v", collab.Community.ID, err)
			return nil, errorx.Unknown
		}

		collaborators = append(
			collaborators,
			convertCollaborator(
				&collab,
				convertCommunity(&collab.Community, int(totalQuests)),
				convertUser(nil, nil, false),
			),
		)
	}

	return &model.GetMyCollabsResponse{Collaborators: collaborators}, nil
}
