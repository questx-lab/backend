package domain

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

type CommunityDomain interface {
	Create(context.Context, *model.CreateCommunityRequest) (*model.CreateCommunityResponse, error)
	GetList(context.Context, *model.GetCommunitiesRequest) (*model.GetCommunitiesResponse, error)
	Get(context.Context, *model.GetCommunityRequest) (*model.GetCommunityResponse, error)
	GetFollowing(context.Context, *model.GetFollowingCommunitiesRequest) (*model.GetFollowingCommunitiesResponse, error)
	UpdateByID(context.Context, *model.UpdateCommunityRequest) (*model.UpdateCommunityResponse, error)
	UpdateDiscord(context.Context, *model.UpdateCommunityDiscordRequest) (*model.UpdateCommunityDiscordResponse, error)
	DeleteByID(context.Context, *model.DeleteCommunityRequest) (*model.DeleteCommunityResponse, error)
	UploadLogo(context.Context, *model.UploadCommunityLogoRequest) (*model.UploadCommunityLogoResponse, error)
	GetMyReferral(context.Context, *model.GetMyReferralRequest) (*model.GetMyReferralResponse, error)
	GetPendingReferral(context.Context, *model.GetPendingReferralRequest) (*model.GetPendingReferralResponse, error)
	ApproveReferral(context.Context, *model.ApproveReferralRequest) (*model.ApproveReferralResponse, error)
}

type communityDomain struct {
	communityRepo         repository.CommunityRepository
	collaboratorRepo      repository.CollaboratorRepository
	userRepo              repository.UserRepository
	communityRoleVerifier *common.CommunityRoleVerifier
	globalRoleVerifier    *common.GlobalRoleVerifier
	discordEndpoint       discord.IEndpoint
	storage               storage.Storage
}

func NewCommunityDomain(
	communityRepo repository.CommunityRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
	discordEndpoint discord.IEndpoint,
	storage storage.Storage,
) CommunityDomain {
	return &communityDomain{
		communityRepo:         communityRepo,
		collaboratorRepo:      collaboratorRepo,
		userRepo:              userRepo,
		discordEndpoint:       discordEndpoint,
		communityRoleVerifier: common.NewCommunityRoleVerifier(collaboratorRepo, userRepo),
		globalRoleVerifier:    common.NewGlobalRoleVerifier(userRepo),
		storage:               storage,
	}
}

func (d *communityDomain) Create(
	ctx context.Context, req *model.CreateCommunityRequest,
) (*model.CreateCommunityResponse, error) {
	referredBy := sql.NullString{Valid: false}
	if req.ReferralCode != "" {
		referralUser, err := d.userRepo.GetByReferralCode(ctx, req.ReferralCode)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Invalid referral code")
			}

			xcontext.Logger(ctx).Errorf("Cannot get referral user: %v", err)
			return nil, errorx.Unknown
		}

		if referralUser.ID == xcontext.RequestUserID(ctx) {
			return nil, errorx.New(errorx.BadRequest, "Cannot refer by yourself")
		}

		referredBy = sql.NullString{Valid: true, String: referralUser.ID}
	}

	userID := xcontext.RequestUserID(ctx)
	proj := &entity.Community{
		Base:               entity.Base{ID: uuid.NewString()},
		Introduction:       []byte(req.Introduction),
		Name:               req.Name,
		WebsiteURL:         req.WebsiteURL,
		DevelopmentStage:   req.DevelopmentStage,
		TeamSize:           req.TeamSize,
		SharedContentTypes: req.SharedContentTypes,
		Twitter:            req.Twitter,
		CreatedBy:          userID,
		ReferredBy:         referredBy,
		ReferralStatus:     entity.ReferralUnclaimable,
	}

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	if err := d.communityRepo.Create(ctx, proj); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create community: %v", err)
		return nil, errorx.Unknown
	}

	err := d.collaboratorRepo.Upsert(ctx, &entity.Collaborator{
		UserID:      userID,
		CommunityID: proj.ID,
		Role:        entity.Owner,
		CreatedBy:   userID,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot assign role owner: %v", err)
		return nil, errorx.Unknown
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.CreateCommunityResponse{ID: proj.ID}, nil
}

func (d *communityDomain) GetList(
	ctx context.Context, req *model.GetCommunitiesRequest,
) (*model.GetCommunitiesResponse, error) {
	if req.Limit == 0 {
		req.Limit = -1
	}

	result, err := d.communityRepo.GetList(ctx, repository.GetListCommunityFilter{
		Q:          req.Q,
		Offset:     req.Offset,
		Limit:      req.Limit,
		ByTrending: req.ByTrending,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get community list: %v", err)
		return nil, errorx.Unknown
	}

	communities := []model.Community{}
	for _, p := range result {
		communities = append(communities, model.Community{
			ID:                 p.ID,
			CreatedAt:          p.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:          p.UpdatedAt.Format(time.RFC3339Nano),
			CreatedBy:          p.CreatedBy,
			ReferredBy:         p.ReferredBy.String,
			Introduction:       string(p.Introduction),
			Name:               p.Name,
			Twitter:            p.Twitter,
			Discord:            p.Discord,
			Followers:          p.Followers,
			TrendingScore:      p.TrendingScore,
			WebsiteURL:         p.WebsiteURL,
			DevelopmentStage:   p.DevelopmentStage,
			TeamSize:           p.TeamSize,
			SharedContentTypes: p.SharedContentTypes,
			LogoURL:            p.LogoPicture,
		})
	}

	return &model.GetCommunitiesResponse{Communities: communities}, nil
}

func (d *communityDomain) Get(ctx context.Context, req *model.GetCommunityRequest) (
	*model.GetCommunityResponse, error) {
	result, err := d.communityRepo.GetByID(ctx, req.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get the community: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetCommunityResponse{Community: model.Community{
		ID:                 result.ID,
		CreatedAt:          result.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:          result.UpdatedAt.Format(time.RFC3339Nano),
		CreatedBy:          result.CreatedBy,
		ReferredBy:         result.ReferredBy.String,
		Introduction:       string(result.Introduction),
		Name:               result.Name,
		Twitter:            result.Twitter,
		Discord:            result.Discord,
		Followers:          result.Followers,
		TrendingScore:      result.TrendingScore,
		WebsiteURL:         result.WebsiteURL,
		DevelopmentStage:   result.DevelopmentStage,
		TeamSize:           result.TeamSize,
		SharedContentTypes: result.SharedContentTypes,
		LogoURL:            result.LogoPicture,
	}}, nil
}

func (d *communityDomain) UpdateByID(
	ctx context.Context, req *model.UpdateCommunityRequest,
) (*model.UpdateCommunityResponse, error) {
	if err := d.communityRoleVerifier.Verify(ctx, req.ID, entity.Owner); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Only owner can update community")
	}

	if req.Name != "" {
		_, err := d.communityRepo.GetByName(ctx, req.Name)
		if err == nil {
			return nil, errorx.New(errorx.AlreadyExists, "The name is already taken by another community")
		}

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			xcontext.Logger(ctx).Errorf("Cannot get community by name: %v", err)
			return nil, errorx.Unknown
		}
	}

	err := d.communityRepo.UpdateByID(ctx, req.ID, entity.Community{
		Name:               req.Name,
		Introduction:       []byte(req.Introduction),
		WebsiteURL:         req.WebsiteURL,
		DevelopmentStage:   req.DevelopmentStage,
		TeamSize:           req.TeamSize,
		SharedContentTypes: req.SharedContentTypes,
		Twitter:            req.Twitter,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update community: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateCommunityResponse{}, nil
}

func (d *communityDomain) UpdateDiscord(
	ctx context.Context, req *model.UpdateCommunityDiscordRequest,
) (*model.UpdateCommunityDiscordResponse, error) {
	if err := d.communityRoleVerifier.Verify(ctx, req.ID, entity.Owner); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Only owner can update discord")
	}

	user, err := d.discordEndpoint.GetMe(ctx, req.AccessToken)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get me discord: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid access token")
	}

	guild, err := d.discordEndpoint.GetGuild(ctx, req.ServerID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get discord server: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid discord server")
	}

	if guild.OwnerID != user.ID {
		return nil, errorx.New(errorx.PermissionDenied, "You are not server's owner")
	}

	err = d.communityRepo.UpdateByID(ctx, req.ID, entity.Community{Discord: guild.ID})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update community: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateCommunityDiscordResponse{}, nil
}

func (d *communityDomain) DeleteByID(
	ctx context.Context, req *model.DeleteCommunityRequest,
) (*model.DeleteCommunityResponse, error) {
	if err := d.communityRoleVerifier.Verify(ctx, req.ID, entity.Owner); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Only owner can delete community")
	}

	if err := d.communityRepo.DeleteByID(ctx, req.ID); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot delete community: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteCommunityResponse{}, nil
}

func (d *communityDomain) GetFollowing(
	ctx context.Context, req *model.GetFollowingCommunitiesRequest,
) (*model.GetFollowingCommunitiesResponse, error) {
	userID := xcontext.RequestUserID(ctx)
	result, err := d.communityRepo.GetFollowingList(ctx, userID, req.Offset, req.Limit)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get community list: %v", err)
		return nil, errorx.Unknown
	}

	communities := []model.Community{}
	for _, p := range result {
		communities = append(communities, model.Community{
			ID:                 p.ID,
			CreatedAt:          p.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:          p.UpdatedAt.Format(time.RFC3339Nano),
			CreatedBy:          p.CreatedBy,
			Name:               p.Name,
			Introduction:       string(p.Introduction),
			Twitter:            p.Twitter,
			Discord:            p.Discord,
			Followers:          p.Followers,
			TrendingScore:      p.TrendingScore,
			WebsiteURL:         p.WebsiteURL,
			DevelopmentStage:   p.DevelopmentStage,
			TeamSize:           p.TeamSize,
			SharedContentTypes: p.SharedContentTypes,
			ReferredBy:         p.ReferredBy.String,
			LogoURL:            p.LogoPicture,
		})
	}

	return &model.GetFollowingCommunitiesResponse{Communities: communities}, nil
}

func (d *communityDomain) UploadLogo(
	ctx context.Context, req *model.UploadCommunityLogoRequest,
) (*model.UploadCommunityLogoResponse, error) {
	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	image, err := common.ProcessImage(ctx, d.storage, "image")
	if err != nil {
		return nil, err
	}

	communityID := xcontext.HTTPRequest(ctx).PostFormValue("community_id")
	if err := d.communityRoleVerifier.Verify(ctx, communityID, entity.Owner); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	community := entity.Community{LogoPicture: image.Url}
	if err := d.communityRepo.UpdateByID(ctx, communityID, community); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update community logo: %v", err)
		return nil, errorx.Unknown
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.UploadCommunityLogoResponse{}, nil
}

func (d *communityDomain) GetMyReferral(
	ctx context.Context, req *model.GetMyReferralRequest,
) (*model.GetMyReferralResponse, error) {
	communities, err := d.communityRepo.GetList(ctx, repository.GetListCommunityFilter{
		ReferredBy: xcontext.RequestUserID(ctx),
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get referral communities: %v", err)
		return nil, errorx.Unknown
	}

	numberOfClaimableCommunities := 0
	numberOfPendingCommunities := 0
	for _, p := range communities {
		if p.ReferralStatus == entity.ReferralClaimable {
			numberOfClaimableCommunities++
		} else if p.ReferralStatus == entity.ReferralPending {
			numberOfPendingCommunities++
		}
	}

	return &model.GetMyReferralResponse{
		TotalClaimableCommunities: numberOfClaimableCommunities,
		TotalPendingCommunities:   numberOfPendingCommunities,
		RewardAmount:              xcontext.Configs(ctx).Quest.InviteCommunityRewardAmount,
	}, nil
}

func (d *communityDomain) GetPendingReferral(
	ctx context.Context, req *model.GetPendingReferralRequest,
) (*model.GetPendingReferralResponse, error) {
	if err := d.globalRoleVerifier.Verify(ctx, entity.GlobalAdminRoles...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denied to get pending referral: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	communities, err := d.communityRepo.GetList(ctx, repository.GetListCommunityFilter{
		ReferralStatus: entity.ReferralPending,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get referral communities: %v", err)
		return nil, errorx.Unknown
	}

	referralCommunities := []model.Community{}
	for _, p := range communities {
		referralCommunities = append(referralCommunities, model.Community{
			ID:                 p.ID,
			CreatedAt:          p.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:          p.UpdatedAt.Format(time.RFC3339Nano),
			CreatedBy:          p.CreatedBy,
			ReferredBy:         p.ReferredBy.String,
			ReferralStatus:     string(p.ReferralStatus),
			Introduction:       string(p.Introduction),
			Name:               p.Name,
			Twitter:            p.Twitter,
			Discord:            p.Discord,
			Followers:          p.Followers,
			TrendingScore:      p.TrendingScore,
			WebsiteURL:         p.WebsiteURL,
			DevelopmentStage:   p.DevelopmentStage,
			TeamSize:           p.TeamSize,
			SharedContentTypes: p.SharedContentTypes,
			LogoURL:            p.LogoPicture,
		})
	}

	return &model.GetPendingReferralResponse{Communities: referralCommunities}, nil
}

func (d *communityDomain) ApproveReferral(
	ctx context.Context, req *model.ApproveReferralRequest,
) (*model.ApproveReferralResponse, error) {
	if err := d.globalRoleVerifier.Verify(ctx, entity.GlobalAdminRoles...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission deined to approve referral: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	communities, err := d.communityRepo.GetByIDs(ctx, req.CommunityIDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get referral communities: %v", err)
		return nil, errorx.Unknown
	}

	for _, p := range communities {
		if p.ReferralStatus != entity.ReferralPending {
			return nil, errorx.New(errorx.BadRequest, "Community %s is not pending status of referral", p.ID)
		}
	}

	err = d.communityRepo.UpdateReferralStatusByIDs(ctx, req.CommunityIDs, entity.ReferralClaimable)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update referral status by ids: %v", err)
		return nil, errorx.Unknown
	}

	return &model.ApproveReferralResponse{}, nil
}
