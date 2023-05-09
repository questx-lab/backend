package domain

import (
	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type ParticipantDomain interface {
	Get(xcontext.Context, *model.GetParticipantRequest) (*model.GetParticipantResponse, error)
	GetList(xcontext.Context, *model.GetListParticipantRequest) (*model.GetListParticipantResponse, error)
}

type participantDomain struct {
	participantRepo repository.ParticipantRepository
	roleVerifier    *common.ProjectRoleVerifier
}

func NewParticipantDomain(
	collaboratorRepo repository.CollaboratorRepository,
	userRepo repository.UserRepository,
	participantRepo repository.ParticipantRepository,
) *participantDomain {
	return &participantDomain{
		participantRepo: participantRepo,
		roleVerifier:    common.NewProjectRoleVerifier(collaboratorRepo, userRepo),
	}
}

func (d *participantDomain) Get(
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
		UserID:      xcontext.GetRequestUserID(ctx),
		Points:      participant.Points,
		InviteCode:  participant.InviteCode,
		InviteCount: participant.InviteCount,
	}

	if participant.InvitedBy.Valid {
		resp.InvitedBy = participant.InvitedBy.String
	}

	return resp, nil
}

func (d *participantDomain) GetList(
	ctx xcontext.Context, req *model.GetListParticipantRequest,
) (*model.GetListParticipantResponse, error) {
	if req.ProjectID == "" {
		return nil, errorx.New(errorx.BadRequest, "Not allow empty project id")
	}

	if err := d.roleVerifier.Verify(ctx, req.ProjectID, entity.ReviewGroup...); err != nil {
		return nil, errorx.New(errorx.PermissionDenied, "Permission denied")
	}

	participants, err := d.participantRepo.GetList(ctx, req.ProjectID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get participant: %v", err)
		return nil, errorx.Unknown
	}

	resp := []model.Participant{}

	for _, p := range participants {
		result := model.Participant{
			UserID:      xcontext.GetRequestUserID(ctx),
			Points:      p.Points,
			InviteCode:  p.InviteCode,
			InviteCount: p.InviteCount,
		}

		if p.InvitedBy.Valid {
			result.InvitedBy = p.InvitedBy.String
		}

		resp = append(resp, result)
	}

	return &model.GetListParticipantResponse{Participants: resp}, nil
}
