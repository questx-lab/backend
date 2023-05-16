package questclaim

import (
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
	ctx xcontext.Context,
	factory Factory,
	data map[string]any,
	needParse bool,
) (*questCondition, error) {
	condition := questCondition{factory: factory}
	err := mapstructure.Decode(data, &condition)
	if err != nil {
		return nil, err
	}

	if needParse {
		if _, err := enum.ToEnum[questConditionOpType](condition.Op); err != nil {
			return nil, err
		}

		dependentQuest, err := factory.questRepo.GetByID(ctx, condition.QuestID)
		if err != nil {
			return nil, err
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

func (c *questCondition) Check(ctx xcontext.Context) (bool, error) {
	targetClaimedQuest, err := c.factory.claimedQuestRepo.GetLast(
		ctx,
		repository.GetLastClaimedQuestFilter{
			UserID:  xcontext.GetRequestUserID(ctx),
			QuestID: c.QuestID,
			Status: []entity.ClaimedQuestStatus{
				entity.Pending,
				entity.Accepted,
				entity.AutoAccepted,
			},
		},
	)

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Logger().Errorf("Cannot get claimed quest: %v", err)
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

// Data Condition
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

func newDateCondition(ctx xcontext.Context, data map[string]any, needParse bool) (*dateCondition, error) {
	condition := dateCondition{}
	err := mapstructure.Decode(data, &condition)
	if err != nil {
		return nil, err
	}

	if needParse {
		_, err := enum.ToEnum[dateConditionOpType](condition.Op)
		if err != nil {
			return nil, err
		}

		_, err = time.Parse("Jan 02 2006", condition.Date)
		if err != nil {
			return nil, err
		}
	}

	return &condition, nil
}

func (c *dateCondition) Statement() string {
	return fmt.Sprintf("You can only claim this quest %s %s", c.Op, c.Date)
}

func (c *dateCondition) Check(xcontext.Context) (bool, error) {
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
