package domain

import (
	"database/sql"
	"errors"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type UserDomain interface {
	GetUser(xcontext.Context, *model.GetUserRequest) (*model.GetUserResponse, error)
	GetInvite(xcontext.Context, *model.GetInviteRequest) (*model.GetInviteResponse, error)
	GetBadges(xcontext.Context, *model.GetBadgesRequest) (*model.GetBadgesResponse, error)
	FollowProject(xcontext.Context, *model.FollowProjectRequest) (*model.FollowProjectResponse, error)
	Assign(xcontext.Context, *model.AssignGlobalRoleRequest) (*model.AssignGlobalRoleResponse, error)
}

type userDomain struct {
	userRepo           repository.UserRepository
	participantRepo    repository.ParticipantRepository
	badgeRepo          repository.BadgeRepo
	badgeManager       *badge.Manager
	globalRoleVerifier *common.GlobalRoleVerifier
}

func NewUserDomain(
	userRepo repository.UserRepository,
	participantRepo repository.ParticipantRepository,
	badgeRepo repository.BadgeRepo,
	badgeManager *badge.Manager,
) UserDomain {
	return &userDomain{
		userRepo:           userRepo,
		participantRepo:    participantRepo,
		badgeRepo:          badgeRepo,
		badgeManager:       badgeManager,
		globalRoleVerifier: common.NewGlobalRoleVerifier(userRepo),
	}
}

func (d *userDomain) GetUser(ctx xcontext.Context, req *model.GetUserRequest) (*model.GetUserResponse, error) {
	user, err := d.userRepo.GetByID(ctx, xcontext.GetRequestUserID(ctx))
	if err != nil {
		ctx.Logger().Errorf("Cannot get user: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetUserResponse{
		ID:      user.ID,
		Address: user.Address.String,
		Name:    user.Name,
		Role:    string(user.Role),
	}, nil
}

func (d *userDomain) GetInvite(
	ctx xcontext.Context, req *model.GetInviteRequest,
) (*model.GetInviteResponse, error) {
	if req.InviteCode == "" {
		return nil, errorx.New(errorx.BadRequest, "Expected a non-empty invite code")
	}

	participant, err := d.participantRepo.GetByReferralCode(ctx, req.InviteCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found invite code")
		}

		ctx.Logger().Errorf("Cannot get participant: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetInviteResponse{
		User: model.User{
			ID:      participant.User.ID,
			Name:    participant.User.Name,
			Address: participant.User.Address.String,
			Role:    string(participant.User.Role),
		},
		Project: model.Project{
			ID:           participant.Project.ID,
			Name:         participant.Project.Name,
			CreatedBy:    participant.Project.CreatedBy,
			Introduction: string(participant.Project.Introduction),
			Twitter:      participant.Project.Twitter,
			Discord:      participant.Project.Discord,
		},
	}, nil
}

func (d *userDomain) GetBadges(
	ctx xcontext.Context, req *model.GetBadgesRequest,
) (*model.GetBadgesResponse, error) {
	badges, err := d.badgeRepo.GetAll(ctx, req.UserID, req.ProjectID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get badges: %v", err)
		return nil, errorx.Unknown
	}

	needUpdate := false
	var clientBadges []model.Badge
	for _, b := range badges {
		clientBadges = append(clientBadges, model.Badge{
			UserID:      b.UserID,
			ProjectID:   b.ProjectID.String,
			Name:        b.Name,
			Level:       b.Level,
			WasNotified: b.WasNotified,
		})

		if !b.WasNotified {
			needUpdate = true
		}
	}

	if needUpdate {
		if err := d.badgeRepo.UpdateNotification(ctx, req.UserID, req.ProjectID); err != nil {
			ctx.Logger().Errorf("Cannot update notification of badge: %v", err)
			return nil, errorx.Unknown
		}
	}

	return &model.GetBadgesResponse{Badges: clientBadges}, nil
}

func (d *userDomain) FollowProject(
	ctx xcontext.Context, req *model.FollowProjectRequest,
) (*model.FollowProjectResponse, error) {
	userID := xcontext.GetRequestUserID(ctx)
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	participant := &entity.Participant{
		UserID:     userID,
		ProjectID:  req.ProjectID,
		InviteCode: crypto.GenerateRandomAlphabet(9),
	}

	ctx.BeginTx()
	defer ctx.RollbackTx()

	if req.InvitedBy != "" {
		participant.InvitedBy = sql.NullString{String: req.InvitedBy, Valid: true}
		err := d.participantRepo.IncreaseInviteCount(ctx, req.InvitedBy, req.ProjectID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Invalid invite user id")
			}

			ctx.Logger().Errorf("Cannot increase invite: %v", err)
			return nil, errorx.Unknown
		}

		err = d.badgeManager.WithBadges(badge.SharpScoutBadgeName).ScanAndGive(ctx, req.InvitedBy, req.ProjectID)
		if err != nil {
			return nil, err
		}
	}

	err := d.participantRepo.Create(ctx, participant)
	if err != nil {
		ctx.Logger().Errorf("Cannot create participant: %v", err)
		return nil, errorx.Unknown
	}

	ctx.CommitTx()
	return &model.FollowProjectResponse{}, nil
}

func (d *userDomain) Assign(
	ctx xcontext.Context, req *model.AssignGlobalRoleRequest,
) (*model.AssignGlobalRoleResponse, error) {
	// user cannot assign by themselves
	if xcontext.GetRequestUserID(ctx) == req.UserID {
		return nil, errorx.New(errorx.PermissionDenied, "Can not assign by yourself")
	}

	role, err := enum.ToEnum[entity.GlobalRole](req.Role)
	if err != nil {
		ctx.Logger().Debugf("Invalid role %s: %v", req.Role, err)
		return nil, errorx.New(errorx.BadRequest, "Invalid role")
	}

	var needRole []entity.GlobalRole
	switch role {
	case entity.RoleSuperAdmin:
		needRole = []entity.GlobalRole{entity.RoleSuperAdmin}
	case entity.RoleAdmin:
		needRole = []entity.GlobalRole{entity.RoleSuperAdmin}
	case entity.RoleUser:
		needRole = entity.GlobalAdminRole
	default:
		return nil, errorx.New(errorx.BadRequest, "Invalid role %s", role)
	}

	// Check permission of the user giving the role against to that role.
	if err = d.globalRoleVerifier.Verify(ctx, needRole...); err != nil {
		ctx.Logger().Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	receivingRoleUser, err := d.userRepo.GetByID(ctx, req.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Logger().Errorf("Cannot get current collaborator of user: %v", err)
		return nil, errorx.Unknown
	}

	if err == nil {
		// Check permission of the user giving role against to the receipent.
		if receivingRoleUser.Role == entity.RoleSuperAdmin || receivingRoleUser.Role == entity.RoleAdmin {
			if err = d.globalRoleVerifier.Verify(ctx, entity.RoleSuperAdmin); err != nil {
				ctx.Logger().Debugf("Permission denied: %v", err)
				return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
			}
		}
	}

	if err := d.userRepo.UpdateByID(ctx, req.UserID, &entity.User{Role: role}); err != nil {
		ctx.Logger().Errorf("Cannot update role of user: %v", err)
		return nil, errorx.Unknown
	}

	return &model.AssignGlobalRoleResponse{}, nil
}
