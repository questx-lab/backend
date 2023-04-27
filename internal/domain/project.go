package domain

import (
	"errors"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"

	"github.com/google/uuid"
)

type ProjectDomain interface {
	Create(ctx xcontext.Context, req *model.CreateProjectRequest) (*model.CreateProjectResponse, error)
	GetMyList(ctx xcontext.Context, req *model.GetMyListProjectRequest) (*model.GetMyListProjectResponse, error)
	GetListByUserID(ctx xcontext.Context, req *model.GetListProjectByUserIDRequest) (*model.GetListProjectByUserIDResponse, error)
	GetList(ctx xcontext.Context, req *model.GetListProjectRequest) (*model.GetListProjectResponse, error)
	GetByID(ctx xcontext.Context, req *model.GetProjectByIDRequest) (*model.GetProjectByIDResponse, error)
	UpdateByID(ctx xcontext.Context, req *model.UpdateProjectByIDRequest) (*model.UpdateProjectByIDResponse, error)
	UpdateDiscord(xcontext.Context, *model.UpdateProjectDiscordRequest) (*model.UpdateProjectDiscordResponse, error)
	DeleteByID(ctx xcontext.Context, req *model.DeleteProjectByIDRequest) (*model.DeleteProjectByIDResponse, error)
}

type projectDomain struct {
	projectRepo      repository.ProjectRepository
	collaboratorRepo repository.CollaboratorRepository
	userRepo         repository.UserRepository
	discordEndpoint  discord.IEndpoint
}

func NewProjectDomain(
	projectRepo repository.ProjectRepository,
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
	discordEndpoint discord.IEndpoint,
) ProjectDomain {
	return &projectDomain{
		projectRepo:      projectRepo,
		collaboratorRepo: collaboratorRepo,
		userRepo:         userRepo,
		discordEndpoint:  discordEndpoint,
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
		Telegram:     req.Telegram,
		CreatedBy:    userID,
	}

	ctx.BeginTx()
	defer ctx.RollbackTx()

	if err := d.projectRepo.Create(ctx, proj); err != nil {
		ctx.Logger().Errorf("Cannot create project: %v", err)
		return nil, errorx.Unknown
	}

	err := d.collaboratorRepo.Create(ctx, &entity.Collaborator{
		Base:      entity.Base{ID: uuid.NewString()},
		UserID:    userID,
		ProjectID: proj.ID,
		Role:      entity.Owner,
	})
	if err != nil {
		ctx.Logger().Errorf("Cannot assign role owner: %v", err)
		return nil, errorx.Unknown
	}

	ctx.CommitTx()

	return &model.CreateProjectResponse{ID: proj.ID}, nil
}

func (d *projectDomain) GetList(ctx xcontext.Context, req *model.GetListProjectRequest) (
	*model.GetListProjectResponse, error) {
	result, err := d.projectRepo.GetList(ctx, req.Offset, req.Limit)
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
			Telegram:     p.Telegram,
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
		Telegram:     result.Telegram,
		Discord:      result.Discord,
	}}, nil
}

func (d *projectDomain) UpdateByID(ctx xcontext.Context, req *model.UpdateProjectByIDRequest) (
	*model.UpdateProjectByIDResponse, error) {
	err := d.projectRepo.UpdateByID(ctx, req.ID, &entity.Project{
		Introduction: []byte(req.Introduction),
		Twitter:      req.Twitter,
		Telegram:     req.Telegram,
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

func (d *projectDomain) DeleteByID(ctx xcontext.Context, req *model.DeleteProjectByIDRequest) (
	*model.DeleteProjectByIDResponse, error) {
	if err := d.projectRepo.DeleteByID(ctx, req.ID); err != nil {
		ctx.Logger().Errorf("Cannot delete project: %v", err)
		return nil, errorx.Unknown
	}

	return &model.DeleteProjectByIDResponse{}, nil
}

func (d *projectDomain) GetMyList(ctx xcontext.Context, req *model.GetMyListProjectRequest) (*model.GetMyListProjectResponse, error) {
	userID := xcontext.GetRequestUserID(ctx)
	result, err := d.projectRepo.GetListByUserID(ctx, userID, req.Offset, req.Limit)
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
			Telegram:     p.Telegram,
			Discord:      p.Discord,
		})
	}

	return &model.GetMyListProjectResponse{Projects: projects}, nil
}

func (d *projectDomain) GetListByUserID(ctx xcontext.Context, req *model.GetListProjectByUserIDRequest) (*model.GetListProjectByUserIDResponse, error) {
	if _, err := d.userRepo.GetByID(ctx, req.UserID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "User not found")
		}
		ctx.Logger().Errorf("Cannot get project list: %v", err)
		return nil, errorx.Unknown
	}
	result, err := d.projectRepo.GetListByUserID(ctx, req.UserID, req.Offset, req.Limit)
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
			Telegram:     p.Telegram,
			Discord:      p.Discord,
		})
	}

	return &model.GetListProjectByUserIDResponse{Projects: projects}, nil
}
