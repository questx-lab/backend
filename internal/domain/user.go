package domain

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
	"gorm.io/gorm"
)

type UserDomain interface {
	GetMe(context.Context, *model.GetMeRequest) (*model.GetMeResponse, error)
	GetUser(ctx context.Context, req *model.GetUserRequest) (*model.GetUserResponse, error)
	Update(context.Context, *model.UpdateUserRequest) (*model.UpdateUserResponse, error)
	GetInvite(context.Context, *model.GetInviteRequest) (*model.GetInviteResponse, error)
	FollowCommunity(context.Context, *model.FollowCommunityRequest) (*model.FollowCommunityResponse, error)
	UnFollowCommunity(context.Context, *model.UnFollowCommunityRequest) (*model.UnFollowCommunityResponse, error)
	Assign(context.Context, *model.AssignGlobalRoleRequest) (*model.AssignGlobalRoleResponse, error)
	UploadAvatar(context.Context, *model.UploadAvatarRequest) (*model.UploadAvatarResponse, error)
}

type userDomain struct {
	userRepo                 repository.UserRepository
	oauth2Repo               repository.OAuth2Repository
	followerRepo             repository.FollowerRepository
	followerRoleRepo         repository.FollowerRoleRepository
	communityRepo            repository.CommunityRepository
	claimedQuestRepo         repository.ClaimedQuestRepository
	globalRoleVerifier       *common.GlobalRoleVerifier
	storage                  storage.Storage
	notificationEngineCaller client.NotificationEngineCaller
	redisClient              xredis.Client
}

func NewUserDomain(
	userRepo repository.UserRepository,
	oauth2Repo repository.OAuth2Repository,
	followerRepo repository.FollowerRepository,
	followerRoleRepo repository.FollowerRoleRepository,
	communityRepo repository.CommunityRepository,
	claimedQuestRepo repository.ClaimedQuestRepository,
	storage storage.Storage,
	notificationEngineCaller client.NotificationEngineCaller,
	redisClient xredis.Client,
) UserDomain {
	return &userDomain{
		userRepo:                 userRepo,
		oauth2Repo:               oauth2Repo,
		followerRepo:             followerRepo,
		followerRoleRepo:         followerRoleRepo,
		communityRepo:            communityRepo,
		claimedQuestRepo:         claimedQuestRepo,
		globalRoleVerifier:       common.NewGlobalRoleVerifier(userRepo),
		storage:                  storage,
		notificationEngineCaller: notificationEngineCaller,
		redisClient:              redisClient,
	}
}

func (d *userDomain) GetMe(ctx context.Context, req *model.GetMeRequest) (*model.GetMeResponse, error) {
	user, err := d.userRepo.GetByID(ctx, xcontext.RequestUserID(ctx))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get user: %v", err)
		return nil, errorx.Unknown
	}

	serviceUsers, err := d.oauth2Repo.GetAllByUserIDs(ctx, user.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get service users: %v", err)
		return nil, errorx.Unknown
	}

	totalCommunites, err := d.followerRepo.Count(
		ctx, repository.StatisticFollowerFilter{UserID: user.ID})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get total joined communities: %v", err)
		return nil, errorx.Unknown
	}

	totalClaimedQuests, err := d.claimedQuestRepo.Count(
		ctx, repository.StatisticClaimedQuestFilter{UserID: user.ID})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get total claimed quests: %v", err)
		return nil, errorx.Unknown
	}

	clientUser := model.ConvertUser(user, serviceUsers, true, "")
	clientUser.TotalCommunities = int(totalCommunites)
	clientUser.TotalClaimedQuests = int(totalClaimedQuests)

	return &model.GetMeResponse{User: clientUser}, nil
}

func (d *userDomain) GetUser(ctx context.Context, req *model.GetUserRequest) (*model.GetUserResponse, error) {
	user, err := d.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found user")
		}

		xcontext.Logger(ctx).Errorf("Cannot get user: %v", err)
		return nil, errorx.Unknown
	}

	totalCommunites, err := d.followerRepo.Count(
		ctx, repository.StatisticFollowerFilter{UserID: req.UserID})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get total joined communities: %v", err)
		return nil, errorx.Unknown
	}

	totalClaimedQuests, err := d.claimedQuestRepo.Count(
		ctx, repository.StatisticClaimedQuestFilter{UserID: req.UserID})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get total claimed quests: %v", err)
		return nil, errorx.Unknown
	}

	clientUser := model.ConvertUser(user, nil, false, "")
	clientUser.TotalCommunities = int(totalCommunites)
	clientUser.TotalClaimedQuests = int(totalClaimedQuests)

	return &model.GetUserResponse{User: clientUser}, nil
}

func (d *userDomain) Update(
	ctx context.Context, req *model.UpdateUserRequest,
) (*model.UpdateUserResponse, error) {
	if err := checkUsername(ctx, req.Name); err != nil {
		return nil, err
	}

	existedUser, err := d.userRepo.GetByName(ctx, req.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		xcontext.Logger(ctx).Errorf("Cannot get user by name: %v", err)
		return nil, errorx.Unknown
	}

	if err == nil && existedUser.ID != xcontext.RequestUserID(ctx) {
		return nil, errorx.New(errorx.AlreadyExists, "This username is already taken")
	}

	oldUser, err := d.userRepo.GetByID(ctx, xcontext.RequestUserID(ctx))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get old user: %v", err)
		return nil, errorx.Unknown
	}

	if !oldUser.IsNewUser && oldUser.Name != req.Name {
		return nil, errorx.New(errorx.Unavailable, "You cannot update your username anymore")
	}

	err = d.userRepo.UpdateByID(ctx, xcontext.RequestUserID(ctx), &entity.User{
		Name:      req.Name,
		IsNewUser: false,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update user: %v", err)
		return nil, errorx.Unknown
	}

	newUser, err := d.userRepo.GetByID(ctx, xcontext.RequestUserID(ctx))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get new user: %v", err)
		return nil, errorx.Unknown
	}

	go func() {
		followers, err := d.followerRepo.GetListByUserID(ctx, newUser.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get my followers info: %v", err)
			return
		}

		for _, f := range followers {
			followerKey := common.RedisKeyFollower(f.CommunityID)
			if exist, err := d.redisClient.Exist(ctx, followerKey); err != nil {
				xcontext.Logger(ctx).Errorf("Cannot check existence of follower key: %v", err)
			} else if exist {
				err := d.redisClient.SRem(
					ctx, followerKey, common.RedisValueFollower(oldUser.Name, oldUser.ID))
				if err != nil {
					xcontext.Logger(ctx).Errorf("Cannot remove user from follower redis: %v", err)
				}

				err = d.redisClient.SAdd(
					ctx, followerKey, common.RedisValueFollower(newUser.Name, newUser.ID))
				if err != nil {
					xcontext.Logger(ctx).Errorf("Cannot add user to follower redis: %v", err)
				}
			}
		}
	}()

	return &model.UpdateUserResponse{User: model.ConvertUser(newUser, nil, true, "")}, nil
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
		User:      model.ConvertUser(&follower.User, nil, false, ""),
		Community: model.ConvertCommunity(&follower.Community, 0),
	}, nil
}

func (d *userDomain) FollowCommunity(
	ctx context.Context, req *model.FollowCommunityRequest,
) (*model.FollowCommunityResponse, error) {
	userID := xcontext.RequestUserID(ctx)
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community handle")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	err = FollowCommunity(
		ctx,
		d.userRepo,
		d.communityRepo,
		d.followerRepo,
		d.followerRoleRepo,
		d.notificationEngineCaller,
		d.redisClient,
		userID, community.ID, req.InvitedBy,
	)
	if err != nil {
		return nil, err
	}

	return &model.FollowCommunityResponse{}, nil
}

func (d *userDomain) UnFollowCommunity(ctx context.Context, req *model.UnFollowCommunityRequest) (*model.UnFollowCommunityResponse, error) {
	userID := xcontext.RequestUserID(ctx)
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community handle")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	follower, err := d.followerRepo.Get(ctx, userID, community.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "User not follow yet")
		}

		xcontext.Logger(ctx).Errorf("Unable to get community: %v", err)
		return nil, errorx.Unknown
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	if follower.InvitedBy.Valid {
		if err := d.followerRepo.DecreaseInviteCount(ctx, follower.InvitedBy.String, follower.CommunityID); err != nil {
			xcontext.Logger(ctx).Errorf("Unable to decrease invite count: %v", err)
			return nil, errorx.Unknown
		}
	}

	if err := d.communityRepo.DecreaseFollowers(ctx, follower.CommunityID); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to decrease follower: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.followerRepo.Delete(ctx, userID, follower.CommunityID); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to delete follower: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.followerRoleRepo.DeleteByUser(ctx, userID, follower.CommunityID); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to delete follower role: %v", err)
		return nil, errorx.Unknown
	}

	ctx = xcontext.WithCommitDBTransaction(ctx)
	followerKey := common.RedisKeyFollower(community.ID)

	if err := d.redisClient.SRem(ctx, followerKey); err != nil {
		xcontext.Logger(ctx).Errorf("Unable to delete follower in cache: %v", err)
	}

	return &model.UnFollowCommunityResponse{}, nil
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

	image, err := common.ProcessFormDataImage(ctx, d.storage, "image")
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
