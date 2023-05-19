package domain

import (
	"database/sql"
	"errors"

	"github.com/questx-lab/backend/internal/domain/badge"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/crypto"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

func followProject(
	ctx xcontext.Context,
	userRepo repository.UserRepository,
	projectRepo repository.ProjectRepository,
	participantRepo repository.ParticipantRepository,
	badgeManager *badge.Manager,
	userID, projectID, invitedBy string,
) error {
	participant := &entity.Participant{
		UserID:     userID,
		ProjectID:  projectID,
		InviteCode: crypto.GenerateRandomAlphabet(9),
	}

	ctx.BeginTx()
	defer ctx.RollbackTx()

	if invitedBy != "" {
		participant.InvitedBy = sql.NullString{String: invitedBy, Valid: true}
		err := participantRepo.IncreaseInviteCount(ctx, invitedBy, projectID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorx.New(errorx.NotFound, "Invalid invite user id")
			}

			ctx.Logger().Errorf("Cannot increase invite: %v", err)
			return errorx.Unknown
		}

		err = badgeManager.
			WithBadges(badge.SharpScoutBadgeName).
			ScanAndGive(ctx, invitedBy, projectID)
		if err != nil {
			return err
		}
	}

	err := participantRepo.Create(ctx, participant)
	if err != nil {
		ctx.Logger().Errorf("Cannot create participant: %v", err)
		return errorx.Unknown
	}

	err = projectRepo.IncreaseFollowers(ctx, projectID)
	if err != nil {
		ctx.Logger().Errorf("Cannot increase followers: %v", err)
		return errorx.Unknown
	}

	project, err := projectRepo.GetByID(ctx, projectID)
	if err != nil {
		ctx.Logger().Errorf("Cannot get project: %v", err)
		return errorx.Unknown
	}

	isUnclaimable := project.ReferralStatus == entity.ReferralUnclaimable
	enoughFollowers := project.Followers >= ctx.Configs().Quest.InviteProjectRequiredFollowers
	if project.ReferredBy.Valid && enoughFollowers && isUnclaimable {
		err = projectRepo.UpdateByID(ctx, project.ID, entity.Project{
			ReferralStatus: entity.ReferralPending,
		})
		if err != nil {
			ctx.Logger().Errorf("Cannot change referral status of project to pending: %v", err)
			return errorx.Unknown
		}
	}

	ctx.CommitTx()
	return nil
}
