package questclaim

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

// Quest Condition
type questConditionOpType string

var (
	isCompleted    = enum.New(questConditionOpType("is_completed"))
	isNotCompleted = enum.New(questConditionOpType("is_not_completed"))
)

type questCondition struct {
	Op         string `mapstructure:"op" structs:"op"`
	QuestID    string `mapstructure:"quest_id" structs:"quest_id"`
	QuestTitle string `mapstructure:"quest_title" structs:"quest_title"`

	factory Factory
}

func newQuestCondition(
	ctx context.Context,
	factory Factory,
	data map[string]any,
	needParse bool,
) (*questCondition, error) {
	condition := questCondition{factory: factory}
	err := mapstructure.Decode(data, &condition)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	if needParse {
		if _, err := enum.ToEnum[questConditionOpType](condition.Op); err != nil {
			xcontext.Logger(ctx).Debugf("Invalid condition op: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid condition op")
		}

		dependentQuest, err := factory.questRepo.GetByID(ctx, condition.QuestID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorx.New(errorx.NotFound, "Not found quest")
			}

			xcontext.Logger(ctx).Warnf("Cannot get quest: %v", err)
			return nil, errorx.Unknown
		}

		condition.QuestTitle = dependentQuest.Title
	}

	return &condition, nil
}

func (c questCondition) Statement() string {
	if c.Op == string(isNotCompleted) {
		return fmt.Sprintf("You can not claim this quest when completed quest %s", c.QuestTitle)
	} else {
		return fmt.Sprintf("Please complete quest %s before claiming this quest", c.QuestTitle)
	}
}

func (c *questCondition) Check(ctx context.Context) (bool, error) {
	targetClaimedQuest, err := c.factory.claimedQuestRepo.GetLast(
		ctx,
		repository.GetLastClaimedQuestFilter{
			UserID:  xcontext.RequestUserID(ctx),
			QuestID: c.QuestID,
			Status: []entity.ClaimedQuestStatus{
				entity.Pending,
				entity.Accepted,
				entity.AutoAccepted,
			},
		},
	)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		xcontext.Logger(ctx).Errorf("Cannot get claimed quest: %v", err)
		return false, errorx.Unknown
	}

	switch questConditionOpType(c.Op) {
	case isCompleted:
		if err != nil {
			return false, nil
		}

		status := targetClaimedQuest.Status
		if status != entity.Accepted && status != entity.AutoAccepted {
			return false, nil
		}

		return true, nil

	case isNotCompleted:
		if err != nil {
			return true, nil
		}

		status := targetClaimedQuest.Status
		if status == entity.Rejected || status == entity.AutoRejected {
			return true, nil
		}

		return false, nil

	default:
		return false, errorx.New(errorx.BadRequest, "Invalid operator of Quest condition")
	}
}

// Date Condition
const ConditionDateFormat = "Jan 02 2006"

type dateConditionOpType string

var (
	dateBefore = enum.New(dateConditionOpType("before"))
	dateAfter  = enum.New(dateConditionOpType("after"))
)

type dateCondition struct {
	Op   string `mapstructure:"op" structs:"op"`
	Date string `mapstructure:"date" structs:"date"`
}

func newDateCondition(ctx context.Context, data map[string]any, needParse bool) (*dateCondition, error) {
	condition := dateCondition{}
	err := mapstructure.Decode(data, &condition)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	if needParse {
		if _, err := enum.ToEnum[dateConditionOpType](condition.Op); err != nil {
			xcontext.Logger(ctx).Debugf("Invalid condition op: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid condition op")
		}

		if _, err = time.Parse(ConditionDateFormat, condition.Date); err != nil {
			xcontext.Logger(ctx).Debugf("Invalid date format: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid date format")
		}
	}

	return &condition, nil
}

func (c *dateCondition) Statement() string {
	return fmt.Sprintf("You can only claim this quest %s %s", c.Op, c.Date)
}

func (c *dateCondition) Check(context.Context) (bool, error) {
	now := time.Now()
	date, err := time.Parse(ConditionDateFormat, c.Date)
	if err != nil {
		return false, err
	}

	switch dateConditionOpType(c.Op) {
	case dateBefore:
		return now.Before(date), nil
	case dateAfter:
		return now.After(date), nil
	default:
		return false, errorx.New(errorx.BadRequest, "Invalid operator of Date condition")
	}
}

// Discord Condition
type discordConditionOpType string

var (
	discordJoined = enum.New(discordConditionOpType("joined"))
	discordMustBe = enum.New(discordConditionOpType("must_be"))
)

type discordCondition struct {
	Op      string `mapstructure:"op" structs:"op"`
	Role    string `mapstructure:"role" structs:"role"`
	RoleID  string `mapstructure:"role_id" structs:"role_id"`
	GuildID string `mapstructure:"guild_id" structs:"guild_id"`

	factory Factory
}

func newDiscordCondition(
	ctx context.Context,
	factory Factory,
	quest entity.Quest,
	data map[string]any,
	needParse bool,
) (*discordCondition, error) {
	condition := discordCondition{factory: factory}
	err := mapstructure.Decode(data, &condition)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode map to struct: %v", err)
		return nil, errorx.Unknown
	}

	if needParse {
		if _, err := enum.ToEnum[discordConditionOpType](condition.Op); err != nil {
			xcontext.Logger(ctx).Debugf("Invalid condition op: %v", err)
			return nil, errorx.New(errorx.BadRequest, "Invalid condition op")
		}

		community, err := factory.communityRepo.GetByID(ctx, quest.CommunityID.String)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot get community: %v", err)
			return nil, errorx.Unknown
		}

		if community.Discord == "" {
			return nil, errorx.New(errorx.Unavailable, "Community has not connected to discord server")
		}
		condition.GuildID = community.Discord

		hasAddBot, err := factory.discordEndpoint.HasAddedBot(ctx, community.Discord)
		if err != nil {
			xcontext.Logger(ctx).Warnf("Cannot call hasAddedBot api: %v", err)
			return nil, errorx.Unknown
		}

		if !hasAddBot {
			return nil, errorx.New(errorx.Unavailable, "Community hasn't added bot to discord server")
		}

		if condition.Role != "" {
			roles, err := factory.discordEndpoint.GetRoles(ctx, community.Discord)
			if err != nil {
				xcontext.Logger(ctx).Debugf("Cannot get roles in discord server: %v", err)
				return nil, errorx.Unknown
			}

			for _, r := range roles {
				if r.Name == condition.Role {
					condition.RoleID = r.ID
					break
				}
			}

			if condition.RoleID == "" {
				return nil, errorx.New(errorx.Unavailable, "Invalid role %s", condition.Role)
			}
		}
	}

	return &condition, nil
}

func (c discordCondition) Statement() string {
	if c.Op == string(discordJoined) {
		return "Can not claim quest when not joined in discord server yet"
	} else {
		return fmt.Sprintf("You must be role %s to claim this quest", c.Role)
	}
}

func (c *discordCondition) Check(ctx context.Context) (bool, error) {
	userDiscordID := c.factory.getRequestServiceUserID(ctx, xcontext.Configs(ctx).Auth.Discord.Name)
	if userDiscordID == "" {
		return false, errorx.New(errorx.Unavailable, "User has not connected to discord")
	}

	user, err := c.factory.discordEndpoint.GetMember(ctx, c.GuildID, userDiscordID)
	if err != nil {
		xcontext.Logger(ctx).Debugf("Not found member: %v", err)
		return false, nil
	}

	switch discordConditionOpType(c.Op) {
	case discordJoined:
		return true, nil

	case discordMustBe:
		for _, role := range user.Roles {
			if role == c.RoleID {
				return true, nil
			}
		}
		return false, nil

	default:
		return false, errorx.New(errorx.BadRequest, "Invalid operator of Quest condition")
	}
}
