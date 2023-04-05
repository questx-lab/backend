package domain

import (
	"database/sql"
	"errors"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type UserDomain interface {
	GetUser(xcontext.Context, *model.GetUserRequest) (*model.GetUserResponse, error)
	GetParticipant(xcontext.Context, *model.GetParticipantRequest) (*model.GetParticipantResponse, error)
	GetReferralInfo(xcontext.Context, *model.GetReferralInfoRequest) (*model.GetReferralInfoResponse, error)
	JoinProject(ctx xcontext.Context, req *model.JoinProjectRequest) (*model.JoinProjectResponse, error)
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
	}, nil
}

func (d *userDomain) GetParticipant(
	ctx xcontext.Context, req *model.GetParticipantRequest,
) (*model.GetParticipantResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	participant, err := d.participantRepo.Get(ctx, xcontext.GetRequestUserID(ctx), req.ProjectID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get participant: %v", err)
		return nil, errorx.Unknown
	}

	resp := &model.GetParticipantResponse{
		Points:        participant.Points,
		ReferralCode:  participant.ReferralCode,
		ReferralCount: participant.ReferralCount,
	}

	if participant.ReferralID.Valid {
		resp.ReferralID = participant.ReferralID.String
	}

	return resp, nil
}

func (d *userDomain) GetReferralInfo(
	ctx xcontext.Context, req *model.GetReferralInfoRequest,
) (*model.GetReferralInfoResponse, error) {
	if req.ReferralCode == "" {
		return nil, errorx.New(errorx.BadRequest, "Expected a non-empty referral code")
	}

	participant, err := d.participantRepo.GetByReferralCode(ctx, req.ReferralCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorx.New(errorx.NotFound, "Not found referral code")
		}

		ctx.Logger().Errorf("Cannot get participant: %v", err)
		return nil, errorx.Unknown
	}

	return &model.GetReferralInfoResponse{
		ReferralID: participant.UserID,
		Project: model.Project{
			ID:   participant.Project.ID,
			Name: participant.Project.Name,
		},
	}, nil
}

func (d *userDomain) JoinProject(
	ctx xcontext.Context, req *model.JoinProjectRequest,
) (*model.JoinProjectResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	participant := &entity.Participant{
		UserID:       xcontext.GetRequestUserID(ctx),
		ProjectID:    req.ProjectID,
		ReferralCode: common.GenerateRandomAlphabet(9),
	}

	if req.ReferralID != "" {
		participant.ReferralID = sql.NullString{String: req.ReferralID, Valid: true}

		err := d.participantRepo.IncreaseReferral(ctx, req.ReferralID, req.ProjectID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Invalid referral id")
			}

			ctx.Logger().Errorf("Cannot increase referral: %v", err)
			return nil, errorx.Unknown
		}
	}

	err := d.participantRepo.Create(ctx, participant)
	if err != nil {
		ctx.Logger().Errorf("Cannot create participant: %v", err)
		return nil, errorx.Unknown
	}

	return &model.JoinProjectResponse{}, nil
}
