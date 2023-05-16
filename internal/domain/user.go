package domain

import (
	"errors"
	"strings"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type UserDomain interface {
	GetUser(xcontext.Context, *model.GetUserRequest) (*model.GetUserResponse, error)
	Update(xcontext.Context, *model.UpdateUserRequest) (*model.UpdateUserResponse, error)
	GetInvite(xcontext.Context, *model.GetInviteRequest) (*model.GetInviteResponse, error)
	GetBadges(xcontext.Context, *model.GetBadgesRequest) (*model.GetBadgesResponse, error)
	GetMyBadges(xcontext.Context, *model.GetMyBadgesRequest) (*model.GetMyBadgesResponse, error)
	FollowProject(xcontext.Context, *model.FollowProjectRequest) (*model.FollowProjectResponse, error)
	Assign(xcontext.Context, *model.AssignGlobalRoleRequest) (*model.AssignGlobalRoleResponse, error)
	UploadAvatar(xcontext.Context, *model.UploadAvatarRequest) (*model.UploadAvatarResponse, error)
}

type userDomain struct {
	userRepo           repository.UserRepository
	oauth2Repo         repository.OAuth2Repository
	participantRepo    repository.ParticipantRepository
	badgeRepo          repository.BadgeRepo
	projectRepo        repository.ProjectRepository
	badgeManager       *badge.Manager
	globalRoleVerifier *common.GlobalRoleVerifier
	storage            storage.Storage
}

func NewUserDomain(
	userRepo repository.UserRepository,
	oauth2Repo repository.OAuth2Repository,
	participantRepo repository.ParticipantRepository,
	badgeRepo repository.BadgeRepo,
	projectRepo repository.ProjectRepository,
	badgeManager *badge.Manager,
	storage storage.Storage,
) UserDomain {
	return &userDomain{
		userRepo:           userRepo,
		oauth2Repo:         oauth2Repo,
		participantRepo:    participantRepo,
		badgeRepo:          badgeRepo,
		projectRepo:        projectRepo,
		badgeManager:       badgeManager,
		globalRoleVerifier: common.NewGlobalRoleVerifier(userRepo),
		storage:            storage,
	}
}

func (d *userDomain) GetUser(ctx xcontext.Context, req *model.GetUserRequest) (*model.GetUserResponse, error) {
	user, err := d.userRepo.GetByID(ctx, xcontext.GetRequestUserID(ctx))
	if err != nil {
		ctx.Logger().Errorf("Cannot get user: %v", err)
		return nil, errorx.Unknown
	}

	serviceUsers, err := d.oauth2Repo.GetAllByUserID(ctx, user.ID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get service users: %v", err)
		return nil, errorx.Unknown
	}

	serviceMap := map[string]string{}
	for _, u := range serviceUsers {
		tag, id, found := strings.Cut(u.ServiceUserID, "_")
		if !found || tag != u.Service {
			return nil, errorx.Unknown
		}

		serviceMap[u.Service] = id
	}

	return &model.GetUserResponse{
		ID:        user.ID,
		Address:   user.Address.String,
		Name:      user.Name,
		Role:      string(user.Role),
		Services:  serviceMap,
		IsNewUser: user.IsNewUser,
	}, nil
}

func (d *userDomain) Update(
	ctx xcontext.Context, req *model.UpdateUserRequest,
) (*model.UpdateUserResponse, error) {
	if req.Name == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow an empty name")
	}

	_, err := d.userRepo.GetByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Logger().Errorf("Cannot get user by name: %v", err)
		return nil, errorx.Unknown
	}

	if err == nil {
		return nil, errorx.New(errorx.AlreadyExists, "This username is already taken")
	}

	err = d.userRepo.UpdateByID(ctx, xcontext.GetRequestUserID(ctx), &entity.User{
		Name:      req.Name,
		IsNewUser: false,
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot update user: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateUserResponse{}, nil
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

	clientBadges := []model.Badge{}
	for _, b := range badges {
		clientBadges = append(clientBadges, model.Badge{
			UserID:    b.UserID,
			ProjectID: b.ProjectID.String,
			Name:      b.Name,
			Level:     b.Level,
		})
	}

	return &model.GetBadgesResponse{Badges: clientBadges}, nil
}

func (d *userDomain) GetMyBadges(
	ctx xcontext.Context, req *model.GetMyBadgesRequest,
) (*model.GetMyBadgesResponse, error) {
	requestUserID := xcontext.GetRequestUserID(ctx)
	badges, err := d.badgeRepo.GetAll(ctx, requestUserID, req.ProjectID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get badges: %v", err)
		return nil, errorx.Unknown
	}

	needUpdate := false
	clientBadges := []model.Badge{}
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
		if err := d.badgeRepo.UpdateNotification(ctx, requestUserID, req.ProjectID); err != nil {
			ctx.Logger().Errorf("Cannot update notification of badge: %v", err)
			return nil, errorx.Unknown
		}
	}

	return &model.GetMyBadgesResponse{Badges: clientBadges}, nil
}

func (d *userDomain) FollowProject(
	ctx xcontext.Context, req *model.FollowProjectRequest,
) (*model.FollowProjectResponse, error) {
	userID := xcontext.GetRequestUserID(ctx)
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	err := followProject(
		ctx,
		d.userRepo,
		d.projectRepo,
		d.participantRepo,
		d.badgeManager,
		userID, req.ProjectID, req.InvitedBy,
	)
	if err != nil {
		return nil, err
	}

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
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found user")
		}

		ctx.Logger().Errorf("Cannot get user: %v", err)
		return nil, errorx.Unknown
	}

	// Check permission of the user giving role against to the receipent.
	if receivingRoleUser.Role == entity.RoleSuperAdmin || receivingRoleUser.Role == entity.RoleAdmin {
		if err = d.globalRoleVerifier.Verify(ctx, entity.RoleSuperAdmin); err != nil {
			ctx.Logger().Debugf("Permission denied: %v", err)
			return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
		}
	}

	if err := d.userRepo.UpdateByID(ctx, req.UserID, &entity.User{Role: role}); err != nil {
		ctx.Logger().Errorf("Cannot update role of user: %v", err)
		return nil, errorx.Unknown
	}

	return &model.AssignGlobalRoleResponse{}, nil
}

func (d *userDomain) UploadAvatar(ctx xcontext.Context, req *model.UploadAvatarRequest) (*model.UploadAvatarResponse, error) {
	ctx.BeginTx()
	defer ctx.RollbackTx()

	images, err := common.ProcessImage(ctx, d.storage, "avatar")
	if err != nil {
		return nil, err
	}

	user := entity.User{ProfilePictures: make(entity.Map)}
	for i, img := range images {
		user.ProfilePictures[common.AvatarSizes[i].String()] = img
	}

	if err := d.userRepo.UpdateByID(ctx, xcontext.GetRequestUserID(ctx), &user); err != nil {
		ctx.Logger().Errorf("Cannot update user avatar: %v", err)
		return nil, errorx.Unknown
	}

	ctx.CommitTx()
	return &model.UploadAvatarResponse{}, nil
}
