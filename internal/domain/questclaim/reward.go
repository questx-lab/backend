package questclaim

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
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

	communityID string
	factory     Factory
}

func newPointReward(
	ctx context.Context,
	quest entity.Quest,
	factory Factory,
	data map[string]any,
) (*pointReward, error) {
	reward := pointReward{factory: factory, communityID: quest.CommunityID.String}
	err := mapstructure.Decode(data, &reward)
	if err != nil {
		return nil, err
	}

	if reward.Points == 0 {
		return nil, errors.New("zero point is not allowed")
	}

	return &reward, nil
}

func (r *pointReward) Give(ctx context.Context, userID, claimedQuestID string) error {
	err := r.factory.followerRepo.IncreaseStat(ctx, userID, r.communityID, int(r.Points), 0)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot increase point to follower: %v", err)
		return errorx.Unknown
	}

	// Update leaderboard.
	for _, rangeType := range entity.UserAggregateRangeList {
		rangeValue, err := dateutil.GetCurrentValueByRange(rangeType)
		if err != nil {
			return err
		}

		if err := r.factory.userAggregateRepo.Upsert(ctx, &entity.UserAggregate{
			CommunityID: r.communityID,
			UserID:      userID,
			Range:       rangeType,
			RangeValue:  rangeValue,
			TotalPoint:  r.Points,
		}); err != nil {
			xcontext.Logger(ctx).Errorf("Cannot increase point to leaderboard: %v", err)
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
	ctx context.Context,
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
		community, err := factory.communityRepo.GetByID(ctx, quest.CommunityID.String)
		if err != nil {
			return nil, err
		}

		if community.Discord == "" {
			return nil, errors.New("community has not connected to discord server")
		}

		reward.GuildID = community.Discord

		roles, err := factory.discordEndpoint.GetRoles(ctx, community.Discord)
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

func (r *discordRoleReward) Give(ctx context.Context, userID, claimedQuestID string) error {
	discordServiceName := xcontext.Configs(ctx).Auth.Discord.Name
	serviceUser, err := r.factory.oauth2Repo.GetByUserID(ctx, discordServiceName, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorx.New(errorx.Unavailable, "User has not connected to discord")
		}

		xcontext.Logger(ctx).Debugf("Cannot get user service id: %v", err)
		return errorx.Unknown
	}

	serviceName, discordID, found := strings.Cut(serviceUser.ServiceUserID, "_")
	if !found || serviceName == discordServiceName {
		return errorx.Unknown
	}

	err = r.factory.discordEndpoint.GiveRole(ctx, r.GuildID, discordID, r.RoleID)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot give role: %v", err)
		return errorx.New(errorx.Internal, "Cannot give role to user")
	}

	return nil
}

// Coin Reward
type coinReward struct {
	Amount    float64 `mapstructure:"amount" structs:"amount"`
	Token     string  `mapstructure:"token" structs:"token"`
	Note      string  `mapstructure:"note" structs:"note"`
	ToAddress string  `mapstructure:"to_address" structs:"to_address"`

	factory Factory
}

func newCoinReward(
	ctx context.Context,
	factory Factory,
	data map[string]any,
	needParse bool,
) (*coinReward, error) {
	reward := coinReward{}
	err := mapstructure.Decode(data, &reward)
	if err != nil {
		return nil, err
	}

	if needParse {
		if reward.Amount <= 0 {
			return nil, errors.New("amount must be a positive")
		}

		if reward.Token == "" {
			return nil, errors.New("not found token")
		}
	}

	reward.factory = factory
	return &reward, nil
}

func (r *coinReward) Give(ctx context.Context, userID, claimedQuestID string) error {
	// TODO: For testing purpose.
	tx := &entity.Transaction{
		Base:   entity.Base{ID: uuid.NewString()},
		UserID: userID,
		Note:   r.Note,
		Status: entity.TransactionPending,
		Token:  r.Token,
		Amount: r.Amount,
	}

	if claimedQuestID != "" {
		tx.ClaimedQuestID = sql.NullString{Valid: true, String: claimedQuestID}
	}

	if r.ToAddress != "" {
		tx.Address = r.ToAddress
	} else {
		user, err := r.factory.userRepo.GetByID(ctx, userID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get user: %v", err)
			return errorx.Unknown
		}

		if !user.Address.Valid {
			return errorx.New(errorx.Unavailable, "User has not connected to wallet yet")
		}

		tx.Address = user.Address.String
	}

	if err := r.factory.transactionRepo.Create(ctx, tx); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create transaction in database: %v", err)
		return errorx.Unknown
	}

	return nil
}
