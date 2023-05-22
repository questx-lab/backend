package domain

import (
	"context"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type FollowerDomain interface {
	Get(context.Context, *model.GetFollowerRequest) (*model.GetFollowerResponse, error)
	GetList(context.Context, *model.GetFollowersRequest) (*model.GetFollowersResponse, error)
}

type followerDomain struct {
	followerRepo repository.FollowerRepository
	roleVerifier *common.CommunityRoleVerifier
}

func NewFollowerDomain(
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
	followerRepo repository.FollowerRepository,
) *followerDomain {
	return &followerDomain{
		followerRepo: followerRepo,
		roleVerifier: common.NewCommunityRoleVerifier(collaboratorRepo, userRepo),
	}
}

func (d *followerDomain) Get(
	ctx context.Context, req *model.GetFollowerRequest,
) (*model.GetFollowerResponse, error) {
	if req.CommunityID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community id")
	}

	follower, err := d.followerRepo.Get(ctx, xcontext.RequestUserID(ctx), req.CommunityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get follower: %v", err)
		return nil, errorx.Unknown
	}

	resp := &model.GetFollowerResponse{
		UserID:      xcontext.RequestUserID(ctx),
		Points:      follower.Points,
		InviteCode:  follower.InviteCode,
		InviteCount: follower.InviteCount,
	}

	if follower.InvitedBy.Valid {
		resp.InvitedBy = follower.InvitedBy.String
	}

	return resp, nil
}

func (d *followerDomain) GetList(
	ctx context.Context, req *model.GetFollowersRequest,
) (*model.GetFollowersResponse, error) {
	if req.CommunityID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty community id")
	}

	if err := d.roleVerifier.Verify(ctx, req.CommunityID, entity.ReviewGroup...); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	followers, err := d.followerRepo.GetList(ctx, req.CommunityID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get followers: %v", err)
		return nil, errorx.Unknown
	}

	resp := []model.Follower{}

	for _, f := range followers {
		result := model.Follower{
			UserID:      xcontext.RequestUserID(ctx),
			Points:      f.Points,
			InviteCode:  f.InviteCode,
			InviteCount: f.InviteCount,
		}

		if f.InvitedBy.Valid {
			result.InvitedBy = f.InvitedBy.String
		}

		resp = append(resp, result)
	}

	return &model.GetFollowersResponse{Followers: resp}, nil
}
