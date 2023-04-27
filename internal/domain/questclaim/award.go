package questclaim

import (
	"errors"
	"strconv"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/api/discord"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
)

// Points Award
type pointAward struct {
	points uint64

	projectID       string
	participantRepo repository.ParticipantRepository
}

func newPointAward(
	ctx xcontext.Context,
	quest entity.Quest,
	participantRepo repository.ParticipantRepository,
	award entity.Award,
) (*pointAward, error) {
	points, err := strconv.ParseUint(award.Value, 10, 0)
	if err != nil {
		return nil, err
	}

	return &pointAward{participantRepo: participantRepo, projectID: quest.ProjectID, points: uint64(points)}, nil
}

func (a *pointAward) Give(ctx xcontext.Context) error {
	return a.participantRepo.IncreasePoint(ctx, xcontext.GetRequestUserID(ctx), a.projectID, a.points)
}

// Discord role Award
type discordRoleAward struct {
	roleID  string
	guildID string

	endpoint discord.IEndpoint
}

func newDiscordRoleAward(
	ctx xcontext.Context,
	quest entity.Quest,
	projectRepo repository.ProjectRepository,
	endpoint discord.IEndpoint,
	award entity.Award,
) (*discordRoleAward, error) {
	project, err := projectRepo.GetByID(ctx, quest.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.Discord == "" {
		return nil, errors.New("project has not connected to discord server")
	}

	roles, err := endpoint.GetRoles(ctx, project.Discord)
	if err != nil {
		return nil, err
	}

	for _, r := range roles {
		if r.Name == award.Value {
			return &discordRoleAward{roleID: r.ID, guildID: project.Discord, endpoint: endpoint}, nil
		}
	}

	return nil, errors.New("invalid role")
}

func (a *discordRoleAward) Give(ctx xcontext.Context) error {
	err := a.endpoint.GiveRole(ctx, a.guildID, a.roleID)
	if err != nil {
		ctx.Logger().Errorf("Cannot give role: %v", err)
		return errorx.New(errorx.Internal, "Cannot give role to user")
	}

	return nil
}
