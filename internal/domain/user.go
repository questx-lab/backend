package domain

import (
	"context"
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
	GetMe(context.Context, *model.GetMeRequest) (*model.GetMeResponse, error)
	Update(context.Context, *model.UpdateUserRequest) (*model.UpdateUserResponse, error)
	GetInvite(context.Context, *model.GetInviteRequest) (*model.GetInviteResponse, error)
	GetBadges(context.Context, *model.GetBadgesRequest) (*model.GetBadgesResponse, error)
	GetMyBadges(context.Context, *model.GetMyBadgesRequest) (*model.GetMyBadgesResponse, error)
	FollowCommunity(context.Context, *model.FollowCommunityRequest) (*model.FollowCommunityResponse, error)
	Assign(context.Context, *model.AssignGlobalRoleRequest) (*model.AssignGlobalRoleResponse, error)
	UploadAvatar(context.Context, *model.UploadAvatarRequest) (*model.UploadAvatarResponse, error)
}

type userDomain struct {
	userRepo           repository.UserRepository
	oauth2Repo         repository.OAuth2Repository
	followerRepo       repository.FollowerRepository
	badgeRepo          repository.BadgeRepo
	communityRepo      repository.CommunityRepository
	badgeManager       *badge.Manager
	globalRoleVerifier *common.GlobalRoleVerifier
	storage            storage.Storage
}

func NewUserDomain(
	userRepo repository.UserRepository,
	oauth2Repo repository.OAuth2Repository,
	followerRepo repository.FollowerRepository,
	badgeRepo repository.BadgeRepo,
	communityRepo repository.CommunityRepository,
	badgeManager *badge.Manager,
	storage storage.Storage,
) UserDomain {
	return &userDomain{
		userRepo:           userRepo,
		oauth2Repo:         oauth2Repo,
		followerRepo:       followerRepo,
		badgeRepo:          badgeRepo,
		communityRepo:      communityRepo,
		badgeManager:       badgeManager,
		globalRoleVerifier: common.NewGlobalRoleVerifier(userRepo),
		storage:            storage,
	}
}

func (d *userDomain) GetMe(ctx context.Context, req *model.GetMeRequest) (*model.GetMeResponse, error) {
	user, err := d.userRepo.GetByID(ctx, xcontext.RequestUserID(ctx))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get user: %v", err)
		return nil, errorx.Unknown
	}

	serviceUsers, err := d.oauth2Repo.GetAllByUserID(ctx, user.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get service users: %v", err)
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

	return &model.GetMeResponse{
		ID:           user.ID,
		Address:      user.Address.String,
		Name:         user.Name,
		Role:         string(user.Role),
		Services:     serviceMap,
		IsNewUser:    user.IsNewUser,
		ReferralCode: user.ReferralCode,
		AvatarURL:    user.ProfilePicture,
	}, nil
}

func (d *userDomain) Update(
	ctx context.Context, req *model.UpdateUserRequest,
) (*model.UpdateUserResponse, error) {
	if req.Name == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow an empty name")
	}

	_, err := d.userRepo.GetByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		xcontext.Logger(ctx).Errorf("Cannot get user by name: %v", err)
		return nil, errorx.Unknown
	}

	if err == nil {
		return nil, errorx.New(errorx.AlreadyExists, "This username is already taken")
	}

	err = d.userRepo.UpdateByID(ctx, xcontext.RequestUserID(ctx), &entity.User{
		Name:      req.Name,
		IsNewUser: false,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update user: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateUserResponse{}, nil
}

func (d *userDomain) GetInvite(
	ctx context.Context, req *model.GetInviteRequest,
) (*model.GetInviteResponse, error) {
	if req.InviteCode == "" {
		return nil, errorx.New(errorx.BadRequest, "Expected a non-empty invite code")
	}

	follower, err := d.followerRepo.GetByReferralCode(ctx, req.InviteCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found invite code")
		}

		xcontext.Logger(ctx).Errorf("Cannot get follower: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetInviteResponse{
		User: model.User{
			ID:        follower.User.ID,
			Name:      follower.User.Name,
			Address:   follower.User.Address.String,
			Role:      string(follower.User.Role),
			AvatarURL: follower.User.ProfilePicture,
		},
		Community: model.Community{
			ID:           follower.Community.ID,
			Name:         follower.Community.Name,
			CreatedBy:    follower.Community.CreatedBy,
			Introduction: string(follower.Community.Introduction),
			Twitter:      follower.Community.Twitter,
			Discord:      follower.Community.Discord,
		},
	}, nil
}

func (d *userDomain) GetBadges(
	ctx context.Context, req *model.GetBadgesRequest,
) (*model.GetBadgesResponse, error) {
	badges, err := d.badgeRepo.GetAll(ctx, req.UserID, req.CommunityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get badges: %v", err)
		return nil, errorx.Unknown
	}

	clientBadges := []model.Badge{}
	for _, b := range badges {
		clientBadges = append(clientBadges, model.Badge{
			UserID:      b.UserID,
			CommunityID: b.CommunityID.String,
			Name:        b.Name,
			Level:       b.Level,
		})
	}

	return &model.GetBadgesResponse{Badges: clientBadges}, nil
}

func (d *userDomain) GetMyBadges(
	ctx context.Context, req *model.GetMyBadgesRequest,
) (*model.GetMyBadgesResponse, error) {
	requestUserID := xcontext.RequestUserID(ctx)
	badges, err := d.badgeRepo.GetAll(ctx, requestUserID, req.CommunityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get badges: %v", err)
		return nil, errorx.Unknown
	}

	needUpdate := false
	clientBadges := []model.Badge{}
	for _, b := range badges {
		clientBadges = append(clientBadges, model.Badge{
			UserID:      b.UserID,
			CommunityID: b.CommunityID.String,
			Name:        b.Name,
			Level:       b.Level,
			WasNotified: b.WasNotified,
		})

		if !b.WasNotified {
			needUpdate = true
		}
	}

	if needUpdate {
		if err := d.badgeRepo.UpdateNotification(ctx, requestUserID, req.CommunityID); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot update notification of badge: %v", err)
			return nil, errorx.Unknown
		}
	}

	return &model.GetMyBadgesResponse{Badges: clientBadges}, nil
}

func (d *userDomain) FollowCommunity(
	ctx context.Context, req *model.FollowCommunityRequest,
) (*model.FollowCommunityResponse, error) {
	userID := xcontext.RequestUserID(ctx)
	if req.CommunityID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community id")
	}

	err := followCommunity(
		ctx,
		d.userRepo,
		d.communityRepo,
		d.followerRepo,
		d.badgeManager,
		userID, req.CommunityID, req.InvitedBy,
	)
	if err != nil {
		return nil, err
	}

	return &model.FollowCommunityResponse{}, nil
}

func (d *userDomain) Assign(
	ctx context.Context, req *model.AssignGlobalRoleRequest,
) (*model.AssignGlobalRoleResponse, error) {
	// user cannot assign by themselves
	if xcontext.RequestUserID(ctx) == req.UserID {
		return nil, errorx.New(errorx.PermissionDenied, "Can not assign by yourself")
	}

	role, err := enum.ToEnum[entity.GlobalRole](req.Role)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Invalid role %s: %v", req.Role, err)
		return nil, errorx.New(errorx.BadRequest, "Invalid role")
	}

	var needRole []entity.GlobalRole
	switch role {
	case entity.RoleSuperAdmin:
		needRole = []entity.GlobalRole{entity.RoleSuperAdmin}
	case entity.RoleAdmin:
		needRole = []entity.GlobalRole{entity.RoleSuperAdmin}
	case entity.RoleUser:
		needRole = entity.GlobalAdminRoles
	default:
		return nil, errorx.New(errorx.BadRequest, "Invalid role %s", role)
	}

	// Check permission of the user giving the role against to that role.
	if err = d.globalRoleVerifier.Verify(ctx, needRole...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	receivingRoleUser, err := d.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found user")
		}

		xcontext.Logger(ctx).Errorf("Cannot get user: %v", err)
		return nil, errorx.Unknown
	}

	// Check permission of the user giving role against to the receipent.
	if receivingRoleUser.Role == entity.RoleSuperAdmin || receivingRoleUser.Role == entity.RoleAdmin {
		if err = d.globalRoleVerifier.Verify(ctx, entity.RoleSuperAdmin); err != nil {
			xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
			return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
		}
	}

	if err := d.userRepo.UpdateByID(ctx, req.UserID, &entity.User{Role: role}); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update role of user: %v", err)
		return nil, errorx.Unknown
	}

	return &model.AssignGlobalRoleResponse{}, nil
}

func (d *userDomain) UploadAvatar(ctx context.Context, req *model.UploadAvatarRequest) (*model.UploadAvatarResponse, error) {
	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	image, err := common.ProcessImage(ctx, d.storage, "image")
	if err != nil {
		return nil, err
	}

	user := entity.User{ProfilePicture: image.Url}
	if err := d.userRepo.UpdateByID(ctx, xcontext.RequestUserID(ctx), &user); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update user avatar: %v", err)
		return nil, errorx.Unknown
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.UploadAvatarResponse{}, nil
}
