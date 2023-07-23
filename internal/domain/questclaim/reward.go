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

type claimedQuestOption struct {
	receivedAddress string
	claimedQuest    *entity.ClaimedQuest
}

func (option *claimedQuestOption) WithClaimedQuest(claimedQuest *entity.ClaimedQuest) {
	option.claimedQuest = claimedQuest
	option.receivedAddress = claimedQuest.WalletAddress
}

func (option *claimedQuestOption) WithWalletAddress(address string) {
	option.receivedAddress = address
}

type referralCommunityOption struct {
	referralCommunity *entity.Community
}

func (option *referralCommunityOption) WithReferralCommunity(referralCommunity *entity.Community) {
	option.referralCommunity = referralCommunity
}

type lotteryWinnerOption struct {
	lotteryWinner *entity.LotteryWinner
}

func (option *lotteryWinnerOption) WithLotteryWinner(winner *entity.LotteryWinner) {
	option.lotteryWinner = winner
}

type commonReward struct {
	claimedQuestOption
	referralCommunityOption
	lotteryWinnerOption
}

func (c *commonReward) getUserID() string {
	switch {
	case c.claimedQuest != nil:
		return c.claimedQuest.UserID
	case c.referralCommunity != nil:
		return c.referralCommunity.ReferredBy.String
	case c.lotteryWinner != nil:
		return c.lotteryWinner.UserID
	}

	return ""
}

// Discord role Reward
type discordRoleReward struct {
	Role    string `mapstructure:"role" structs:"role"`
	RoleID  string `mapstructure:"role_id" structs:"role_id"`
	GuildID string `mapstructure:"guild_id" structs:"guild_id"`

	factory Factory
	commonReward
}

func newDiscordRoleReward(
	ctx context.Context,
	communityID string,
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
		community, err := factory.communityRepo.GetByID(ctx, communityID)
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

func (r *discordRoleReward) Give(ctx context.Context) error {
	var userID = r.getUserID()
	if userID == "" {
		xcontext.Logger(ctx).Errorf("Not found user to give role")
		return errorx.Unknown
	}

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
	if !found || serviceName != discordServiceName {
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
	Amount       float64 `mapstructure:"amount" structs:"amount"`
	Chain        string  `mapstructure:"chain" structs:"chain"`
	TokenID      string  `mapstructure:"token_id" structs:"token_id"`
	TokenSymbol  string  `mapstructure:"token_symbol" structs:"token_symbol"`
	TokenAddress string  `mapstructure:"token_address" structs:"token_address"`

	factory Factory
	commonReward
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
		if reward.Chain == "" {
			return nil, errorx.New(errorx.BadRequest, "Not found chain")
		}

		if reward.TokenAddress == "" {
			return nil, errorx.New(errorx.BadRequest, "Not found token")
		}

		if reward.Amount <= 0 {
			return nil, errorx.New(errorx.BadRequest, "Amount must be a positive")
		}

		if err = factory.blockchainRepo.Check(ctx, reward.Chain); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Got an unsupported chain %s", reward.Chain)
			}

			xcontext.Logger(ctx).Errorf("Cannot check chain: %v", err)
			return nil, errorx.Unknown
		}

		tokenInfo, err := factory.blockchainRepo.GetToken(ctx, reward.Chain, reward.TokenAddress)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Got an unsupported token %s on chain %s",
					reward.TokenAddress, reward.Chain)
			}

			xcontext.Logger(ctx).Errorf("Cannot get token: %v", err)
			return nil, errorx.Unknown
		}

		reward.TokenSymbol = tokenInfo.Symbol
		reward.TokenID = tokenInfo.ID
	}

	reward.factory = factory
	return &reward, nil
}

func (r *coinReward) Give(ctx context.Context) error {
	payreward := &entity.PayReward{
		Base:          entity.Base{ID: uuid.NewString()},
		TokenID:       r.TokenID,
		Amount:        r.Amount,
		ToUserID:      r.getUserID(),
		TransactionID: sql.NullString{Valid: false}, // pending for processing at blockchain service.
	}

	if payreward.ToUserID == "" {
		xcontext.Logger(ctx).Errorf("Not found user to pay reward")
		return errorx.Unknown
	}

	// Determine the reason to give this pay reward.
	switch {
	case r.claimedQuest != nil:
		quest, err := r.factory.questRepo.GetByID(ctx, r.claimedQuest.QuestID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get quest when give reward: %v", err)
			return errorx.Unknown
		}

		if r.claimedQuest.Status != entity.Accepted && r.claimedQuest.Status != entity.AutoAccepted {
			return errorx.New(errorx.Unavailable, "Claimed quest is not accepted")
		}

		payreward.ClaimedQuestID = sql.NullString{Valid: true, String: r.claimedQuest.ID}
		payreward.FromCommunityID = sql.NullString{Valid: true, String: quest.CommunityID.String}

	case r.referralCommunity != nil:
		payreward.ReferralCommunityID = sql.NullString{Valid: true, String: r.referralCommunity.ID}
		payreward.FromCommunityID = sql.NullString{} // From our platform

	case r.lotteryWinner != nil:
		prize, err := r.factory.lotteryRepo.GetPrizeByID(ctx, r.lotteryWinner.LotteryPrizeID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get prize: %v", err)
			return errorx.Unknown
		}

		event, err := r.factory.lotteryRepo.GetEventByID(ctx, prize.LotteryEventID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get lottery event: %v", err)
			return errorx.Unknown
		}

		payreward.LotteryWinnerID = sql.NullString{Valid: true, String: r.lotteryWinner.ID}
		payreward.FromCommunityID = sql.NullString{Valid: true, String: event.CommunityID}
	}

	// Check if user provided a customized wallet address, if not, use the
	// linked address.
	if r.receivedAddress == "" {
		user, err := r.factory.userRepo.GetByID(ctx, payreward.ToUserID)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get user when give reward: %v", err)
			return errorx.Unknown
		}

		payreward.ToAddress = user.WalletAddress.String
	} else {
		payreward.ToAddress = r.receivedAddress
	}

	if payreward.ToAddress == "" {
		return errorx.New(errorx.Unavailable,
			"User must choose a wallet address or link to a wallet to receive the reward")
	}

	if err := r.factory.payRewardRepo.Create(ctx, payreward); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot create transaction in database: %v", err)
		return errorx.Unknown
	}

	return nil
}
