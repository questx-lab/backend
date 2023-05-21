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

type ProjectDomain interface {
	Create(context.Context, *model.CreateProjectRequest) (*model.CreateProjectResponse, error)
	GetList(context.Context, *model.GetListProjectRequest) (*model.GetListProjectResponse, error)
	GetFollowing(context.Context, *model.GetFollowingProjectRequest) (*model.GetFollowingProjectResponse, error)
	GetByID(context.Context, *model.GetProjectByIDRequest) (*model.GetProjectByIDResponse, error)
	UpdateByID(context.Context, *model.UpdateProjectByIDRequest) (*model.UpdateProjectByIDResponse, error)
	UpdateDiscord(context.Context, *model.UpdateProjectDiscordRequest) (*model.UpdateProjectDiscordResponse, error)
	DeleteByID(context.Context, *model.DeleteProjectByIDRequest) (*model.DeleteProjectByIDResponse, error)
	UploadLogo(context.Context, *model.UploadProjectLogoRequest) (*model.UploadProjectLogoResponse, error)
	GetMyReferral(context.Context, *model.GetMyReferralRequest) (*model.GetMyReferralResponse, error)
	GetPendingReferral(context.Context, *model.GetPendingReferralProjectsRequest) (*model.GetPendingReferralProjectsResponse, error)
	ApproveReferral(context.Context, *model.ApproveReferralProjectsRequest) (*model.ApproveReferralProjectsResponse, error)
}

type projectDomain struct {
	projectRepo         repository.ProjectRepository
	collaboratorRepo    repository.CollaboratorRepository
	userRepo            repository.UserRepository
	projectRoleVerifier *common.ProjectRoleVerifier
	globalRoleVerifier  *common.GlobalRoleVerifier
	discordEndpoint     discord.IEndpoint
	storage             storage.Storage
}

func NewProjectDomain(
	projectRepo repository.ProjectRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
	discordEndpoint discord.IEndpoint,
	storage storage.Storage,
) ProjectDomain {
	return &projectDomain{
		projectRepo:         projectRepo,
		collaboratorRepo:    collaboratorRepo,
		userRepo:            userRepo,
		discordEndpoint:     discordEndpoint,
		projectRoleVerifier: common.NewProjectRoleVerifier(collaboratorRepo, userRepo),
		globalRoleVerifier:  common.NewGlobalRoleVerifier(userRepo),
		storage:             storage,
	}
}

func (d *projectDomain) Create(
	ctx context.Context, req *model.CreateProjectRequest,
) (*model.CreateProjectResponse, error) {
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
	proj := &entity.Project{
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

	if err := d.projectRepo.Create(ctx, proj); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create project: %v", err)
		return nil, errorx.Unknown
	}

	err := d.collaboratorRepo.Upsert(ctx, &entity.Collaborator{
		UserID:    userID,
		ProjectID: proj.ID,
		Role:      entity.Owner,
		CreatedBy: userID,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot assign role owner: %v", err)
		return nil, errorx.Unknown
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.CreateProjectResponse{ID: proj.ID}, nil
}

func (d *projectDomain) GetList(
	ctx context.Context, req *model.GetListProjectRequest,
) (*model.GetListProjectResponse, error) {
	if req.Limit == 0 {
		req.Limit = -1
	}

	result, err := d.projectRepo.GetList(ctx, repository.GetListProjectFilter{
		Q:               req.Q,
		Offset:          req.Offset,
		Limit:           req.Limit,
		OrderByTrending: req.OrderByTrending,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get project list: %v", err)
		return nil, errorx.Unknown
	}

	projects := []model.Project{}
	for _, p := range result {
		projects = append(projects, model.Project{
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
		})
	}

	return &model.GetListProjectResponse{Projects: projects}, nil
}

func (d *projectDomain) GetByID(ctx context.Context, req *model.GetProjectByIDRequest) (
	*model.GetProjectByIDResponse, error) {
	result, err := d.projectRepo.GetByID(ctx, req.ID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get the project: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetProjectByIDResponse{Project: model.Project{
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
	}}, nil
}

func (d *projectDomain) UpdateByID(
	ctx context.Context, req *model.UpdateProjectByIDRequest,
) (*model.UpdateProjectByIDResponse, error) {
	if err := d.projectRoleVerifier.Verify(ctx, req.ID, entity.Owner); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Only owner can update project")
	}

	if req.Name != "" {
		_, err := d.projectRepo.GetByName(ctx, req.Name)
		if err == nil {
			return nil, errorx.New(errorx.AlreadyExists, "The name is already taken by another project")
		}

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			xcontext.Logger(ctx).Errorf("Cannot get project by name: %v", err)
			return nil, errorx.Unknown
		}
	}

	err := d.projectRepo.UpdateByID(ctx, req.ID, entity.Project{
		Name:               req.Name,
		Introduction:       []byte(req.Introduction),
		WebsiteURL:         req.WebsiteURL,
		DevelopmentStage:   req.DevelopmentStage,
		TeamSize:           req.TeamSize,
		SharedContentTypes: req.SharedContentTypes,
		Twitter:            req.Twitter,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update project: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateProjectByIDResponse{}, nil
}

func (d *projectDomain) UpdateDiscord(
	ctx context.Context, req *model.UpdateProjectDiscordRequest,
) (*model.UpdateProjectDiscordResponse, error) {
	if err := d.projectRoleVerifier.Verify(ctx, req.ID, entity.Owner); err != nil {
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

	err = d.projectRepo.UpdateByID(ctx, req.ID, entity.Project{Discord: guild.ID})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update project: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateProjectDiscordResponse{}, nil
}

func (d *projectDomain) DeleteByID(
	ctx context.Context, req *model.DeleteProjectByIDRequest,
) (*model.DeleteProjectByIDResponse, error) {
	if err := d.projectRoleVerifier.Verify(ctx, req.ID, entity.Owner); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Only owner can delete project")
	}

	if err := d.projectRepo.DeleteByID(ctx, req.ID); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot delete project: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteProjectByIDResponse{}, nil
}

func (d *projectDomain) GetFollowing(
	ctx context.Context, req *model.GetFollowingProjectRequest,
) (*model.GetFollowingProjectResponse, error) {
	userID := xcontext.RequestUserID(ctx)
	result, err := d.projectRepo.GetFollowingList(ctx, userID, req.Offset, req.Limit)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get project list: %v", err)
		return nil, errorx.Unknown
	}

	projects := []model.Project{}
	for _, p := range result {
		projects = append(projects, model.Project{
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
		})
	}

	return &model.GetFollowingProjectResponse{Projects: projects}, nil
}

func (d *projectDomain) UploadLogo(
	ctx context.Context, req *model.UploadProjectLogoRequest,
) (*model.UploadProjectLogoResponse, error) {
	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	images, err := common.ProcessImage(ctx, d.storage, "logo")
	if err != nil {
		return nil, err
	}

	project := entity.Project{LogoPictures: make(entity.Map)}
	for i, img := range images {
		project.LogoPictures[common.AvatarSizes[i].String()] = img
	}

	if err := d.projectRepo.UpdateByID(ctx, xcontext.RequestUserID(ctx), project); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update project logo: %v", err)
		return nil, errorx.Unknown
	}

	xcontext.WithCommitDBTransaction(ctx)
	return &model.UploadProjectLogoResponse{}, nil
}

func (d *projectDomain) GetMyReferral(
	ctx context.Context, req *model.GetMyReferralRequest,
) (*model.GetMyReferralResponse, error) {
	projects, err := d.projectRepo.GetList(ctx, repository.GetListProjectFilter{
		ReferredBy: xcontext.RequestUserID(ctx),
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get referral projects: %v", err)
		return nil, errorx.Unknown
	}

	numberOfClaimableProjects := 0
	numberOfPendingProjects := 0
	for _, p := range projects {
		if p.ReferralStatus == entity.ReferralClaimable {
			numberOfClaimableProjects++
		} else if p.ReferralStatus == entity.ReferralPending {
			numberOfPendingProjects++
		}
	}

	return &model.GetMyReferralResponse{
		TotalClaimableProjects: numberOfClaimableProjects,
		TotalPendingProjects:   numberOfPendingProjects,
		RewardAmount:           xcontext.Configs(ctx).Quest.InviteProjectRewardAmount,
	}, nil
}

func (d *projectDomain) GetPendingReferral(
	ctx context.Context, req *model.GetPendingReferralProjectsRequest,
) (*model.GetPendingReferralProjectsResponse, error) {
	if err := d.globalRoleVerifier.Verify(ctx, entity.GlobalAdminRoles...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denined to get pending referral: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	projects, err := d.projectRepo.GetList(ctx, repository.GetListProjectFilter{
		ReferralStatus: entity.ReferralPending,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get referral projects: %v", err)
		return nil, errorx.Unknown
	}

	referralProjects := []model.Project{}
	for _, p := range projects {
		referralProjects = append(referralProjects, model.Project{
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
		})
	}

	return &model.GetPendingReferralProjectsResponse{Projects: referralProjects}, nil
}

func (d *projectDomain) ApproveReferral(
	ctx context.Context, req *model.ApproveReferralProjectsRequest,
) (*model.ApproveReferralProjectsResponse, error) {
	if err := d.globalRoleVerifier.Verify(ctx, entity.GlobalAdminRoles...); err != nil {
		xcontext.Logger(ctx).Debugf("Permission denined to approve referral: %v", err)
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	projects, err := d.projectRepo.GetByIDs(ctx, req.ProjectIDs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get referral projects: %v", err)
		return nil, errorx.Unknown
	}

	for _, p := range projects {
		if p.ReferralStatus != entity.ReferralPending {
			return nil, errorx.New(errorx.BadRequest, "Project %s is not pending status of referral", p.ID)
		}
	}

	err = d.projectRepo.UpdateReferralStatusByIDs(ctx, req.ProjectIDs, entity.ReferralClaimable)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot update referral status by ids: %v", err)
		return nil, errorx.Unknown
	}

	return &model.ApproveReferralProjectsResponse{}, nil
}
