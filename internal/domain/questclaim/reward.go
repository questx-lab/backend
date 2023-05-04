package questclaim

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/dateutil"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

// Points Reward
type pointReward struct {
	Points uint64 `mapstructure:"points" structs:"points"`

	projectID string
	factory   Factory
}

func newPointReward(
	ctx xcontext.Context,
	quest entity.Quest,
	factory Factory,
	data map[string]any,
) (*pointReward, error) {
	reward := pointReward{factory: factory, projectID: quest.ProjectID}
	err := mapstructure.Decode(data, &reward)
	if err != nil {
		return nil, err
	}

	if reward.Points == 0 {
		return nil, errors.New("zero point is not allowed")
	}

	return &reward, nil
}

func (a *pointReward) Give(ctx xcontext.Context, userID string) error {
	err := a.factory.participantRepo.IncreasePoint(ctx, userID, a.projectID, a.Points)
	if err != nil {
		ctx.Logger().Errorf("Cannot increase point to participant: %v", err)
		return errorx.Unknown
	}

	// Update leaderboard.
	for _, r := range entity.UserAggregateRangeList {
		rangeValue, err := dateutil.GetCurrentValueByRange(r)
		if err != nil {
			return err
		}

		if err := a.factory.userAggregateRepo.Upsert(ctx, &entity.UserAggregate{
			ProjectID:  a.projectID,
			UserID:     userID,
			Range:      r,
			RangeValue: rangeValue,
			TotalPoint: a.Points,
		}); err != nil {
			ctx.Logger().Errorf("Cannot increase point to leaderboard: %v", err)
			return errorx.Unknown
		}
	}

	return nil
}

// Discord role Reward
type discordRoleReward struct {
	Role    string `mapstructure:"role" structs:"role"`
	RoleID  string `mapstructure:"role_id" structs:"role_id"`
	GuildID string `mapstructure:"guild_id" structs:"guild_id"`

	factory Factory
}

func newDiscordRoleReward(
	ctx xcontext.Context,
	quest entity.Quest,
	factory Factory,
	data map[string]any,
	needParse bool,
) (*discordRoleReward, error) {
	reward := discordRoleReward{factory: factory}
	err := mapstructure.Decode(data, &reward)
	if err != nil {
		return nil, err
	}

	if needParse {
		project, err := factory.projectRepo.GetByID(ctx, quest.ProjectID)
		if err != nil {
			return nil, err
		}

		if project.Discord == "" {
			return nil, errors.New("project has not connected to discord server")
		}

		reward.GuildID = project.Discord

		roles, err := factory.discordEndpoint.GetRoles(ctx, project.Discord)
		if err != nil {
			return nil, err
		}

		for _, r := range roles {
			if r.Name == reward.Role {
				reward.RoleID = r.ID
				break
			}
		}

		if reward.RoleID == "" {
			return nil, fmt.Errorf("invalid role %s", reward.Role)
		}
	}

	return &reward, nil
}

func (a *discordRoleReward) Give(ctx xcontext.Context, userID string) error {
	serviceUser, err := a.factory.oauth2Repo.GetByUserID(
		ctx, ctx.Configs().Auth.Discord.Name, xcontext.GetRequestUserID(ctx))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorx.New(errorx.Unavailable, "User has not connected to discord")
		}

		ctx.Logger().Debugf("Cannot get user service id: %v", err)
		return errorx.Unknown
	}

	serviceName, discordID, found := strings.Cut(serviceUser.ServiceUserID, "_")
	if !found || serviceName == ctx.Configs().Auth.Discord.Name {
		return errorx.Unknown
	}

	err = a.factory.discordEndpoint.GiveRole(ctx, a.GuildID, discordID, a.RoleID)
	if err != nil {
		ctx.Logger().Errorf("Cannot give role: %v", err)
		return errorx.New(errorx.Internal, "Cannot give role to user")
	}

	return nil
}
