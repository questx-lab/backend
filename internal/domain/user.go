package domain

import (
	"database/sql"
	"errors"

	"github.com/questx-lab/backend/internal/domain/badge"
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
	GetBadges(xcontext.Context, *model.GetBadgesRequest) (*model.GetBadgesResponse, error)
	FollowProject(ctx xcontext.Context, req *model.FollowProjectRequest) (*model.FollowProjectResponse, error)
}

type userDomain struct {
	userRepo        repository.UserRepository
	participantRepo repository.ParticipantRepository
	badgeRepo       repository.BadgeRepo
	badgeManager    *badge.Manager
}

func NewUserDomain(
	userRepo repository.UserRepository,
	participantRepo repository.ParticipantRepository,
	badgeRepo repository.BadgeRepo,
	badgeManager *badge.Manager,
) UserDomain {
	return &userDomain{
		userRepo:        userRepo,
		participantRepo: participantRepo,
		badgeRepo:       badgeRepo,
		badgeManager:    badgeManager,
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
		User: model.User{
			ID:      participant.User.ID,
			Name:    participant.User.Name,
			Address: participant.User.Address,
			Role:    string(participant.User.Role),
		},
		Project: model.Project{
			ID:           participant.Project.ID,
			Name:         participant.Project.Name,
			CreatedBy:    participant.Project.CreatedBy,
			Introduction: string(participant.Project.Introduction),
			Twitter:      participant.Project.Twitter,
			Discord:      participant.Project.Discord,
		},
	}, nil
}

func (d *userDomain) GetBadges(
	ctx xcontext.Context, req *model.GetBadgesRequest,
) (*model.GetBadgesResponse, error) {
	badges, err := d.badgeRepo.GetAll(ctx, req.UserID, req.ProjectID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get badges: %v", err)
		return nil, errorx.Unknown
	}

	needUpdate := false
	var clientBadges []model.Badge
	for _, b := range badges {
		clientBadges = append(clientBadges, model.Badge{
			UserID:      b.UserID,
			ProjectID:   b.ProjectID.String,
			Name:        b.Name,
			Level:       b.Level,
			WasNotified: b.WasNotified,
		})

		if !b.WasNotified {
			needUpdate = true
		}
	}

	if needUpdate {
		if err := d.badgeRepo.UpdateNotification(ctx, req.UserID, req.ProjectID); err != nil {
			ctx.Logger().Errorf("Cannot update notification of badge: %v", err)
			return nil, errorx.Unknown
		}
	}

	return &model.GetBadgesResponse{Badges: clientBadges}, nil
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
				return nil, errorx.New(errorx.NotFound, "Invalid invite user id")
			}

			ctx.Logger().Errorf("Cannot increase invite: %v", err)
			return nil, errorx.Unknown
		}

		err = d.badgeManager.WithBadges(badge.SharpScoutBadgeName).ScanAndGive(ctx, req.InvitedBy, req.ProjectID)
		if err != nil {
			ctx.Logger().Errorf("Cannot scan and give badge: %v", err)
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
