package domain

import (
	"database/sql"
	"errors"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type UserDomain interface {
	GetUser(xcontext.Context, *model.GetUserRequest) (*model.GetUserResponse, error)
	GetInvite(xcontext.Context, *model.GetInviteRequest) (*model.GetInviteResponse, error)
	FollowProject(ctx xcontext.Context, req *model.FollowProjectRequest) (*model.FollowProjectResponse, error)
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
	user, err := d.userRepo.GetByID(ctx, xcontext.GetRequestUserID(ctx))
	if err != nil {
		ctx.Logger().Errorf("Cannot get user: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetUserResponse{
		ID:      user.ID,
		Address: user.Address,
		Name:    user.Name,
		Role:    string(user.Role),
	}, nil
}

func (d *userDomain) GetInvite(
	ctx xcontext.Context, req *model.GetInviteRequest,
) (*model.GetInviteResponse, error) {
	if req.InviteCode == "" {
		return nil, errorx.New(errorx.BadRequest, "Expected a non-empty invite code")
	}

	participant, err := d.participantRepo.GetByReferralCode(ctx, req.InviteCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found invite code")
		}

		ctx.Logger().Errorf("Cannot get participant: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetInviteResponse{
		InvitedBy: participant.UserID,
		Project: model.Project{
			ID:   participant.Project.ID,
			Name: participant.Project.Name,
		},
	}, nil
}

func (d *userDomain) FollowProject(
	ctx xcontext.Context, req *model.FollowProjectRequest,
) (*model.FollowProjectResponse, error) {
	userID := xcontext.GetRequestUserID(ctx)
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	participant := &entity.Participant{
		UserID:     userID,
		ProjectID:  req.ProjectID,
		InviteCode: crypto.GenerateRandomAlphabet(9),
	}

	ctx.BeginTx()
	defer ctx.RollbackTx()

	if req.InvitedBy != "" {
		participant.InvitedBy = sql.NullString{String: req.InvitedBy, Valid: true}

		err := d.participantRepo.IncreaseInviteCount(ctx, req.InvitedBy, req.ProjectID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Invalid invite id")
			}

			ctx.Logger().Errorf("Cannot increase invite: %v", err)
			return nil, errorx.Unknown
		}
	}

	err := d.participantRepo.Create(ctx, participant)
	if err != nil {
		ctx.Logger().Errorf("Cannot create participant: %v", err)
		return nil, errorx.Unknown
	}

	ctx.CommitTx()
	return &model.FollowProjectResponse{}, nil
}
