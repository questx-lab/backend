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
	roleVerifier     *common.CommunityRoleVerifier
}

func NewFollowerDomain(
	followerRepo repository.FollowerRepository,
	followerRoleRepo repository.FollowerRoleRepository,
	communityRepo repository.CommunityRepository,
	roleRepo repository.RoleRepository,
	roleVerifier *common.CommunityRoleVerifier,
) *followerDomain {
	return &followerDomain{
		followerRepo:     followerRepo,
		followerRoleRepo: followerRoleRepo,
		communityRepo:    communityRepo,
		roleRepo:         roleRepo,
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

		clientRoles = append(clientRoles, convertRole(role))
	}

	resp := model.GetFollowerResponse(convertFollower(
		follower, clientRoles, convertUser(nil, nil, false), convertCommunity(community, 0)))

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

			clientRoles = append(clientRoles, convertRole(role))
		}

		clientFollowers = append(clientFollowers, convertFollower(
			&f, clientRoles, convertUser(nil, nil, false), convertCommunity(&community, 0)))
	}

	return &model.GetAllMyFollowersResponse{Followers: clientFollowers}, nil
}

func (d *followerDomain) GetByCommunityID(
	ctx context.Context, req *model.GetFollowersRequest,
) (*model.GetFollowersResponse, error) {
	if req.CommunityHandle == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community id")
	}

	community, err := d.communityRepo.GetByHandle(ctx, req.CommunityHandle)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found community")
		}

		xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
		return nil, errorx.Unknown
	}

	if err := d.roleVerifier.Verify(ctx, community.ID); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}
	var (
		followers []entity.Follower
	)
	if req.Q != "" {
		followers, err = d.followerRepo.SearchByCommunityID(ctx, community.ID, req.Q)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get followers: %v", err)
			return nil, errorx.Unknown
		}
	} else {
		followers, err = d.followerRepo.GetListByCommunityID(ctx, community.ID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get followers: %v", err)
			return nil, errorx.Unknown
		}
	}

	userIDs := []string{}
	for i := range followers {
		userIDs = append(userIDs, followers[i].UserID)
	}

	followerRoles, err := d.followerRoleRepo.GetMultipleUser(ctx, community.ID, userIDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get follower roles: %v", err)
		return nil, errorx.Unknown
	}

	roleMap := map[string]entity.Role{}
	roleByUserMap := map[string][]string{}
	for _, fr := range followerRoles {
		if fr.RoleID == entity.UserBaseRole && req.IgnoreUserRole {
			continue
		}
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

	communityModel := model.Community{Handle: req.CommunityHandle}
	resp := []model.Follower{}
	for _, f := range followers {
		roleIDs, ok := roleByUserMap[f.UserID]
		if !ok || len(roleIDs) == 0 {
			if req.IgnoreUserRole {
				continue
			}
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

			clientRoles = append(clientRoles, convertRole(&role))
		}

		resp = append(resp, convertFollower(
			&f, clientRoles, convertUser(nil, nil, false), communityModel))
	}

	return &model.GetFollowersResponse{Followers: resp}, nil
}
