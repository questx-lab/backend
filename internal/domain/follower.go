package domain

import (
	"context"
	"errors"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type FollowerDomain interface {
	Get(context.Context, *model.GetFollowerRequest) (*model.GetFollowerResponse, error)
	GetByUserID(context.Context, *model.GetAllMyFollowersRequest) (*model.GetAllMyFollowersResponse, error)
	GetByCommunityID(context.Context, *model.GetFollowersRequest) (*model.GetFollowersResponse, error)
}

type followerDomain struct {
	followerRepo     repository.FollowerRepository
	followerRoleRepo repository.FollowerRoleRepository
	communityRepo    repository.CommunityRepository
	roleRepo         repository.RoleRepository
	questRepo        repository.QuestRepository
	roleVerifier     *common.CommunityRoleVerifier
	userRepo         repository.UserRepository
}

func NewFollowerDomain(
	followerRepo repository.FollowerRepository,
	followerRoleRepo repository.FollowerRoleRepository,
	communityRepo repository.CommunityRepository,
	roleRepo repository.RoleRepository,
	userRepo repository.UserRepository,
	questRepo repository.QuestRepository,
	roleVerifier *common.CommunityRoleVerifier,
) *followerDomain {
	return &followerDomain{
		followerRepo:     followerRepo,
		followerRoleRepo: followerRoleRepo,
		communityRepo:    communityRepo,
		roleRepo:         roleRepo,
		userRepo:         userRepo,
		questRepo:        questRepo,
		roleVerifier:     roleVerifier,
	}
}

func (d *followerDomain) Get(
	ctx context.Context, req *model.GetFollowerRequest,
) (*model.GetFollowerResponse, error) {
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

	follower, err := d.followerRepo.Get(ctx, xcontext.RequestUserID(ctx), community.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get follower: %v", err)
		return nil, errorx.Unknown
	}

	followerRoles, err := d.followerRoleRepo.Get(ctx, follower.UserID, follower.CommunityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get follower roles: %v", err)
		return nil, errorx.Unknown
	}

	clientRoles := []model.Role{}
	for _, followerRole := range followerRoles {
		role, err := d.roleRepo.GetByID(ctx, followerRole.RoleID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get role %s: %v", followerRole.RoleID, err)
			return nil, errorx.Unknown
		}

		clientRoles = append(clientRoles, model.ConvertRole(role))
	}

	resp := model.GetFollowerResponse(model.ConvertFollower(
		follower, clientRoles, model.ConvertUser(nil, nil, false), model.ConvertCommunity(community, 0)))

	return &resp, nil
}

func (d *followerDomain) GetByUserID(
	ctx context.Context, req *model.GetAllMyFollowersRequest,
) (*model.GetAllMyFollowersResponse, error) {
	requestUserID := xcontext.RequestUserID(ctx)
	followers, err := d.followerRepo.GetListByUserID(ctx, requestUserID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get followers: %v", err)
		return nil, errorx.Unknown
	}

	communities, err := d.communityRepo.GetFollowingList(ctx, requestUserID, 0, -1)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get community list: %v", err)
		return nil, errorx.Unknown
	}

	communityMap := map[string]entity.Community{}
	for _, c := range communities {
		communityMap[c.ID] = c
	}

	clientFollowers := []model.Follower{}
	for _, f := range followers {
		community, ok := communityMap[f.CommunityID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Cannot find community for follower %s", f.UserID)
			return nil, errorx.Unknown
		}

		followerRoles, err := d.followerRoleRepo.Get(ctx, f.UserID, f.CommunityID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get follower roles: %v", err)
			return nil, errorx.Unknown
		}

		clientRoles := []model.Role{}
		for _, followerRole := range followerRoles {
			role, err := d.roleRepo.GetByID(ctx, followerRole.RoleID)
			if err != nil {
				xcontext.Logger(ctx).Errorf("Cannot get role %s: %v", followerRole.RoleID, err)
				return nil, errorx.Unknown
			}

			clientRoles = append(clientRoles, model.ConvertRole(role))
		}

		totalQuests, err := d.questRepo.Count(
			ctx, repository.StatisticQuestFilter{CommunityID: f.CommunityID})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot count quest of community %s: %v", f.CommunityID, err)
			return nil, errorx.Unknown
		}

		clientFollowers = append(clientFollowers, model.ConvertFollower(
			&f, clientRoles, model.ConvertUser(nil, nil, false), model.ConvertCommunity(&community, int(totalQuests))))
	}

	return &model.GetAllMyFollowersResponse{Followers: clientFollowers}, nil
}

func (d *followerDomain) GetByCommunityID(
	ctx context.Context, req *model.GetFollowersRequest,
) (*model.GetFollowersResponse, error) {
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community id")
	}

	if req.IgnoreUserRole {
		// In case ignore_user_role is enabled, we don't need a pagination
		// because almost all users are ignored.
		req.Offset = 0
		req.Limit = -1
	} else {
		if req.Offset < 0 {
			return nil, errorx.New(errorx.BadRequest, "Not allow negative offset")
		}

		apiCfg := xcontext.Configs(ctx).ApiServer
		if req.Limit == 0 {
			req.Limit = apiCfg.DefaultLimit
		}

		if req.Limit < 0 {
			return nil, errorx.New(errorx.BadRequest, "Limit must be positive")
		}

		if req.Limit > apiCfg.MaxLimit {
			return nil, errorx.New(errorx.BadRequest, "Exceed the maximum of limit (%d)", apiCfg.MaxLimit)
		}
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	followers, err := d.followerRepo.GetListByCommunityID(ctx, repository.GetListFollowerFilter{
		CommunityID:    community.ID,
		Q:              req.Q,
		IgnoreUserRole: req.IgnoreUserRole,
		Offset:         req.Offset,
		Limit:          req.Limit,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get followers: %v", err)
		return nil, errorx.Unknown
	}

	userIDs := []string{}
	for i := range followers {
		userIDs = append(userIDs, followers[i].UserID)
	}

	followerRoles, err := d.followerRoleRepo.GetByCommunityAndUserIDs(ctx, community.ID, userIDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get follower roles: %v", err)
		return nil, errorx.Unknown
	}

	roleMap := map[string]entity.Role{}
	roleByUserMap := map[string][]string{}
	for _, fr := range followerRoles {
		roleMap[fr.RoleID] = entity.Role{}
		roleByUserMap[fr.UserID] = append(roleByUserMap[fr.UserID], fr.RoleID)
	}

	roles, err := d.roleRepo.GetByIDs(ctx, common.MapKeys(roleMap))
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get roles: %v", err)
		return nil, errorx.Unknown
	}

	for _, r := range roles {
		roleMap[r.ID] = r
	}

	users, err := d.userRepo.GetByIDs(ctx, userIDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get users: %v", err)
		return nil, errorx.Unknown
	}

	userMap := make(map[string]*entity.User)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	communityModel := model.Community{Handle: req.CommunityHandle}
	resp := []model.Follower{}
	for _, f := range followers {
		roleIDs, ok := roleByUserMap[f.UserID]
		if !ok {
			xcontext.Logger(ctx).Errorf("Cannot get follower roles of user %s", f.UserID)
			return nil, errorx.Unknown
		}

		clientRoles := []model.Role{}
		for _, roleID := range roleIDs {
			role, ok := roleMap[roleID]
			if !ok {
				xcontext.Logger(ctx).Errorf("Cannot get role %s", roleID)
				return nil, errorx.Unknown
			}

			clientRoles = append(clientRoles, model.ConvertRole(&role))
		}
		resp = append(resp, model.ConvertFollower(
			&f, clientRoles, model.ConvertUser(userMap[f.UserID], nil, false), communityModel))
	}

	return &model.GetFollowersResponse{Followers: resp}, nil
}
