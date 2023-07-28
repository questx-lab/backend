package domain

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type RoleDomain interface {
	CreateRole(context.Context, *model.CreateRoleRequest) (*model.CreateRoleResponse, error)
	UpdateRole(context.Context, *model.UpdateRoleRequest) (*model.UpdateRoleResponse, error)
	DeleteRole(context.Context, *model.DeleteRoleRequest) (*model.UpdateRoleResponse, error)
	GetRoles(context.Context, *model.GetRolesRequest) (*model.GetRolesResponse, error)
}

type roleDomain struct {
	roleRepo      repository.RoleRepository
	communityRepo repository.CommunityRepository
	roleVerifier  *common.CommunityRoleVerifier
}

func NewRoleDomain(roleRepo repository.RoleRepository,
	communityRepo repository.CommunityRepository,
	roleVerifier *common.CommunityRoleVerifier,
) RoleDomain {
	return &roleDomain{
		roleRepo:      roleRepo,
		communityRepo: communityRepo,
		roleVerifier:  roleVerifier,
	}
}

func (d *roleDomain) CreateRole(ctx context.Context, req *model.CreateRoleRequest) (*model.CreateRoleResponse, error) {
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community handle")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown

	}
	communityID := community.ID
	if err := d.roleVerifier.Verify(ctx, communityID); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	latestPriorityRole, err := d.roleRepo.GetLatestPriorityByCommunityID(ctx, communityID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorx.Unknown
	}

	priority := 0
	if latestPriorityRole != nil {
		priority = latestPriorityRole.Priority
	}

	if err := d.roleRepo.Create(ctx, &entity.Role{
		Base: entity.Base{
			ID: uuid.NewString(),
		},
		CommunityID: sql.NullString{
			String: communityID,
			Valid:  true,
		},
		Permissions: uint64(req.Permissions),
		Name:        req.Name,
		Priority:    priority + 1,
		Color:       req.Color,
	}); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to create role for community: %v", err)
		return nil, errorx.Unknown
	}

	return &model.CreateRoleResponse{}, nil
}

func (d *roleDomain) UpdateRole(ctx context.Context, req *model.UpdateRoleRequest) (*model.UpdateRoleResponse, error) {
	if slices.Contains([]string{entity.OwnerBaseRole, entity.UserBaseRole}, req.RoleID) {
		return nil, errorx.New(errorx.PermissionDenied, "Unable to update base role")
	}

	role, err := d.roleRepo.GetByID(ctx, req.RoleID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get role by id: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, role.CommunityID.String, role.ID); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}
	if err := d.roleRepo.UpdateByID(ctx, role.ID, &entity.Role{
		Name:        req.Name,
		Permissions: uint64(req.Permissions),
		Priority:    req.Priority,
		Color:       req.Color,
	}); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	return &model.UpdateRoleResponse{}, nil
}

func (d *roleDomain) GetRoles(ctx context.Context, req *model.GetRolesRequest) (*model.GetRolesResponse, error) {
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community handle")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown

	}
	communityID := community.ID
	roles, err := d.roleRepo.GetByCommunityID(ctx, communityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get roles by community id: %v", err)
		return nil, errorx.Unknown
	}

	baseRoles, err := d.roleRepo.GetByIDs(ctx, []string{entity.OwnerBaseRole, entity.UserBaseRole})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get roles by community id: %v", err)
		return nil, errorx.Unknown
	}
	respRoles := []model.Role{}
	for _, role := range append(roles, baseRoles...) {
		respRoles = append(respRoles, convertRole(&role))
	}

	return &model.GetRolesResponse{
		Roles: respRoles,
	}, nil
}

func (d *roleDomain) DeleteRole(ctx context.Context, req *model.DeleteRoleRequest) (*model.UpdateRoleResponse, error) {
	if slices.Contains([]string{entity.OwnerBaseRole, entity.UserBaseRole}, req.RoleID) {
		return nil, errorx.New(errorx.PermissionDenied, "Unable to update base role")
	}

	role, err := d.roleRepo.GetByID(ctx, req.RoleID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Unable to get role by id: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, role.CommunityID.String, role.ID); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}
	if err := d.roleRepo.DeleteByID(ctx, role.ID); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to delete role: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleRepo.UpdatePriorityByDelete(ctx, role.CommunityID.String, role.Priority); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to delete role: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateRoleResponse{}, nil
}
