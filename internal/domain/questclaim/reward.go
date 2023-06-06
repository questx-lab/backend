package questclaim

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

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
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	if needParse {
		community, err := factory.communityRepo.GetByID(ctx, quest.CommunityID.String)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}

		if community.Discord == "" {
			return nil, errorx.New(errorx.Unavailable, "Community has not connected to discord server")
		}
		reward.GuildID = community.Discord

		hasAddBot, err := factory.discordEndpoint.HasAddedBot(ctx, community.Discord)
		if err != nil {
			xcontext.Logger(ctx).Warnf("Cannot call hasAddedBot api: %v", err)
			return nil, errorx.Unknown
		}

		if !hasAddBot {
			return nil, errorx.New(errorx.Unavailable, "Community hasn't added bot to discord server")
		}

		roles, err := factory.discordEndpoint.GetRoles(ctx, community.Discord)
		if err != nil {
			xcontext.Logger(ctx).Debugf("Cannot get roles in discord server: %v", err)
			return nil, errorx.Unknown
		}

		for _, r := range roles {
			if r.Name == reward.Role {
				reward.RoleID = r.ID
				break
			}
		}

		if reward.RoleID == "" {
			return nil, errorx.New(errorx.Unavailable, "Invalid role %s", reward.Role)
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
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	if needParse {
		if reward.Amount <= 0 {
			return nil, errorx.New(errorx.BadRequest, "Amount must be a positive")
		}

		if reward.Token == "" {
			return nil, errorx.New(errorx.NotFound, "Not found token")
		}
	}

	reward.factory = factory
	return &reward, nil
}

func (r *coinReward) Give(ctx context.Context, userID, claimedQuestID string) error {
	// TODO: For testing purpose.
	tx := &entity.PayReward{
		Base:   entity.Base{ID: uuid.NewString()},
		UserID: userID,
		Note:   r.Note,
		Status: entity.PayRewardPending,
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

		if !user.WalletAddress.Valid {
			return errorx.New(errorx.Unavailable, "User has not connected to wallet yet")
		}

		tx.Address = user.WalletAddress.String
	}

	if err := r.factory.payRewardRepo.Create(ctx, tx); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create transaction in database: %v", err)
		return errorx.Unknown
	}

	return nil
}
