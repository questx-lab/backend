package domain

import (
	"time"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/google/uuid"
)

type ProjectDomain interface {
	Create(xcontext.Context, *model.CreateProjectRequest) (*model.CreateProjectResponse, error)
	GetList(xcontext.Context, *model.GetListProjectRequest) (*model.GetListProjectResponse, error)
	GetFollowing(xcontext.Context, *model.GetFollowingProjectRequest) (*model.GetFollowingProjectResponse, error)
	GetByID(xcontext.Context, *model.GetProjectByIDRequest) (*model.GetProjectByIDResponse, error)
	UpdateByID(xcontext.Context, *model.UpdateProjectByIDRequest) (*model.UpdateProjectByIDResponse, error)
	UpdateDiscord(xcontext.Context, *model.UpdateProjectDiscordRequest) (*model.UpdateProjectDiscordResponse, error)
	DeleteByID(xcontext.Context, *model.DeleteProjectByIDRequest) (*model.DeleteProjectByIDResponse, error)
}

type projectDomain struct {
	projectRepo         repository.ProjectRepository
	collaboratorRepo    repository.CollaboratorRepository
	userRepo            repository.UserRepository
	projectRoleVerifier *common.ProjectRoleVerifier
	discordEndpoint     discord.IEndpoint
}

func NewProjectDomain(
	projectRepo repository.ProjectRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
	discordEndpoint discord.IEndpoint,
) ProjectDomain {
	return &projectDomain{
		projectRepo:         projectRepo,
		collaboratorRepo:    collaboratorRepo,
		userRepo:            userRepo,
		discordEndpoint:     discordEndpoint,
		projectRoleVerifier: common.NewProjectRoleVerifier(collaboratorRepo, userRepo),
	}
}

func (d *projectDomain) Create(ctx xcontext.Context, req *model.CreateProjectRequest) (
	*model.CreateProjectResponse, error) {
	userID := xcontext.GetRequestUserID(ctx)
	proj := &entity.Project{
		Base:         entity.Base{ID: uuid.NewString()},
		Introduction: []byte(req.Introduction),
		Name:         req.Name,
		Twitter:      req.Twitter,
		CreatedBy:    userID,
	}

	ctx.BeginTx()
	defer ctx.RollbackTx()

	if err := d.projectRepo.Create(ctx, proj); err != nil {
		ctx.Logger().Errorf("Cannot create project: %v", err)
		return nil, errorx.Unknown
	}

	err := d.collaboratorRepo.Upsert(ctx, &entity.Collaborator{
		UserID:    userID,
		ProjectID: proj.ID,
		Role:      entity.Owner,
		CreatedBy: userID,
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot assign role owner: %v", err)
		return nil, errorx.Unknown
	}

	ctx.CommitTx()

	return &model.CreateProjectResponse{ID: proj.ID}, nil
}

func (d *projectDomain) GetList(
	ctx xcontext.Context, req *model.GetListProjectRequest,
) (*model.GetListProjectResponse, error) {
	if req.Limit == 0 {
		req.Limit = -1
	}

	result, err := d.projectRepo.GetList(ctx, repository.GetListProjectFilter{
		Q:      req.Q,
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot get project list: %v", err)
		return nil, errorx.Unknown
	}

	projects := []model.Project{}
	for _, p := range result {
		projects = append(projects, model.Project{
			ID:           p.ID,
			CreatedAt:    p.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:    p.UpdatedAt.Format(time.RFC3339Nano),
			CreatedBy:    p.CreatedBy,
			Introduction: string(p.Introduction),
			Name:         p.Name,
			Twitter:      p.Twitter,
			Discord:      p.Discord,
		})
	}

	return &model.GetListProjectResponse{Projects: projects}, nil
}

func (d *projectDomain) GetByID(ctx xcontext.Context, req *model.GetProjectByIDRequest) (
	*model.GetProjectByIDResponse, error) {
	result, err := d.projectRepo.GetByID(ctx, req.ID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get the project: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetProjectByIDResponse{Project: model.Project{
		ID:           result.ID,
		CreatedAt:    result.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:    result.UpdatedAt.Format(time.RFC3339Nano),
		CreatedBy:    result.CreatedBy,
		Introduction: string(result.Introduction),
		Name:         result.Name,
		Twitter:      result.Twitter,
		Discord:      result.Discord,
	}}, nil
}

func (d *projectDomain) UpdateByID(
	ctx xcontext.Context, req *model.UpdateProjectByIDRequest,
) (*model.UpdateProjectByIDResponse, error) {
	if err := d.projectRoleVerifier.Verify(ctx, req.ID, entity.Owner); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Only owner can update project")
	}

	err := d.projectRepo.UpdateByID(ctx, req.ID, &entity.Project{
		Introduction: []byte(req.Introduction),
		Twitter:      req.Twitter,
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot update project: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateProjectByIDResponse{}, nil
}

func (d *projectDomain) UpdateDiscord(
	ctx xcontext.Context, req *model.UpdateProjectDiscordRequest,
) (*model.UpdateProjectDiscordResponse, error) {
	if err := d.projectRoleVerifier.Verify(ctx, req.ID, entity.Owner); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Only owner can update discord")
	}

	user, err := d.discordEndpoint.GetMe(ctx, req.AccessToken)
	if err != nil {
		ctx.Logger().Errorf("Cannot get me discord: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid access token")
	}

	guild, err := d.discordEndpoint.GetGuild(ctx, req.ServerID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get discord server: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid discord server")
	}

	if guild.OwnerID != user.ID {
		return nil, errorx.New(errorx.PermissionDenied, "You are not server's owner")
	}

	err = d.projectRepo.UpdateByID(ctx, req.ID, &entity.Project{Discord: guild.ID})
	if err != nil {
		ctx.Logger().Errorf("Cannot update project: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UpdateProjectDiscordResponse{}, nil
}

func (d *projectDomain) DeleteByID(
	ctx xcontext.Context, req *model.DeleteProjectByIDRequest,
) (*model.DeleteProjectByIDResponse, error) {
	if err := d.projectRoleVerifier.Verify(ctx, req.ID, entity.Owner); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Only owner can delete project")
	}

	if err := d.projectRepo.DeleteByID(ctx, req.ID); err != nil {
		ctx.Logger().Errorf("Cannot delete project: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteProjectByIDResponse{}, nil
}

func (d *projectDomain) GetFollowing(
	ctx xcontext.Context, req *model.GetFollowingProjectRequest,
) (*model.GetFollowingProjectResponse, error) {
	userID := xcontext.GetRequestUserID(ctx)
	result, err := d.projectRepo.GetFollowingList(ctx, userID, req.Offset, req.Limit)
	if err != nil {
		ctx.Logger().Errorf("Cannot get project list: %v", err)
		return nil, errorx.Unknown
	}

	projects := []model.Project{}
	for _, p := range result {
		projects = append(projects, model.Project{
			ID:           p.ID,
			CreatedAt:    p.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:    p.UpdatedAt.Format(time.RFC3339Nano),
			CreatedBy:    p.CreatedBy,
			Name:         p.Name,
			Introduction: string(p.Introduction),
			Twitter:      p.Twitter,
			Discord:      p.Discord,
		})
	}

	return &model.GetFollowingProjectResponse{Projects: projects}, nil
}
