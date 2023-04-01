package domain

import (
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type UserDomain interface {
	JoinProject(ctx xcontext.Context, req *model.JoinProjectRequest) (*model.JoinProjectResponse, error)
	GetUser(xcontext.Context, *model.GetUserRequest) (*model.GetUserResponse, error)
	GetPoints(xcontext.Context, *model.GetPointsRequest) (*model.GetPointsResponse, error)
}

type userDomain struct {
	userRepo        repository.UserRepository
	participantRepo repository.ParticipantRepository
}

func NewUserDomain(
	userRepo repository.UserRepository,
	participantRepo repository.ParticipantRepository,
) UserDomain {
	return &userDomain{
		userRepo:        userRepo,
		participantRepo: participantRepo,
	}
}

func (d *userDomain) GetUser(ctx xcontext.Context, req *model.GetUserRequest) (*model.GetUserResponse, error) {
	user, err := d.userRepo.GetByID(ctx, ctx.GetUserID())
	if err != nil {
		ctx.Logger().Errorf("Cannot get user: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetUserResponse{
		ID:      user.ID,
		Address: user.Address,
		Name:    user.Name,
	}, nil
}

func (d *userDomain) JoinProject(
	ctx xcontext.Context, req *model.JoinProjectRequest,
) (*model.JoinProjectResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	err := d.participantRepo.Create(ctx, ctx.GetUserID(), req.ProjectID)
	if err != nil {
		ctx.Logger().Errorf("Cannot create participant: %v", err)
		return nil, errorx.Unknown
	}

	return &model.JoinProjectResponse{}, nil
}

func (d *userDomain) GetPoints(
	ctx xcontext.Context, req *model.GetPointsRequest,
) (*model.GetPointsResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	participant, err := d.participantRepo.Get(ctx, ctx.GetUserID(), req.ProjectID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get participant: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetPointsResponse{Points: participant.Points}, nil
}
