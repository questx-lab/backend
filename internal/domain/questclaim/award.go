package questclaim

import (
	"strconv"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

// Points Award
type pointAward struct {
	participantRepo repository.ParticipantRepository
	points          uint64
}

func newPointAward(
	ctx xcontext.Context,
	participantRepo repository.ParticipantRepository,
	award entity.Award,
) (*pointAward, error) {
	points, err := strconv.ParseUint(award.Value, 10, 0)
	if err != nil {
		return nil, err
	}

	return &pointAward{participantRepo: participantRepo, points: uint64(points)}, nil
}

func (a *pointAward) Give(ctx xcontext.Context, projectID string) error {
	return a.participantRepo.Increase(ctx, xcontext.GetRequestUserID(ctx), projectID, a.points)
}

// Discord role Award
type discordRoleAward struct {
	role string
}

func newDiscordRoleAward(ctx xcontext.Context, award entity.Award) (*discordRoleAward, error) {
	// TODO: Need to check if role existed.
	return &discordRoleAward{role: award.Value}, nil
}

func (a *discordRoleAward) Give(ctx xcontext.Context, projectID string) error {
	return errorx.New(errorx.NotImplemented, "not implemented discord role award")
}
