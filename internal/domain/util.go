package domain

import (
	"context"
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
	ctx context.Context,
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

	ctx = xcontext.WithDBTransaction(ctx)
	defer xcontext.WithRollbackDBTransaction(ctx)

	if invitedBy != "" {
		participant.InvitedBy = sql.NullString{String: invitedBy, Valid: true}
		err := participantRepo.IncreaseInviteCount(ctx, invitedBy, projectID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorx.New(errorx.NotFound, "Invalid invite user id")
			}

			xcontext.Logger(ctx).Errorf("Cannot increase invite: %v", err)
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
		xcontext.Logger(ctx).Errorf("Cannot create participant: %v", err)
		return errorx.Unknown
	}

	err = projectRepo.IncreaseFollowers(ctx, projectID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot increase followers: %v", err)
		return errorx.Unknown
	}

	project, err := projectRepo.GetByID(ctx, projectID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get project: %v", err)
		return errorx.Unknown
	}

	isUnclaimable := project.ReferralStatus == entity.ReferralUnclaimable
	enoughFollowers := project.Followers >= xcontext.Configs(ctx).Quest.InviteProjectRequiredFollowers
	if project.ReferredBy.Valid && enoughFollowers && isUnclaimable {
		err = projectRepo.UpdateByID(ctx, project.ID, entity.Project{
			ReferralStatus: entity.ReferralPending,
		})
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot change referral status of project to pending: %v", err)
			return errorx.Unknown
		}
	}

	ctx = xcontext.WithCommitDBTransaction(ctx)
	return nil
}
