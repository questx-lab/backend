package questutil

import (
	"errors"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/enum"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
	"gorm.io/gorm"
)

// Quest Condition
type questConditionOpType string

var (
	isCompleted    = enum.New(questConditionOpType("is completed"))
	isNotCompleted = enum.New(questConditionOpType("is not completed"))
)

type questCondition struct {
	claimedQuestRepo repository.ClaimedQuestRepository

	op      questConditionOpType
	questID string
}

func newQuestCondition(
	ctx router.Context,
	condition entity.Condition,
	claimedQuestRepo repository.ClaimedQuestRepository,
	questRepo repository.QuestRepository,
) (*questCondition, error) {
	op, err := enum.ToEnum[questConditionOpType](condition.Op)
	if err != nil {
		return nil, err
	}

	_, err = questRepo.GetByID(ctx, condition.Value)
	if err != nil {
		return nil, err
	}

	return &questCondition{
		claimedQuestRepo: claimedQuestRepo,
		op:               op, questID: condition.Value,
	}, nil
}

func (c *questCondition) Check(ctx router.Context) (bool, error) {
	targetClaimedQuest, err := c.claimedQuestRepo.GetLastPendingOrAccepted(ctx, ctx.GetUserID(), c.questID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Logger().Errorf("Cannot get claimed quest: %v", err)
		return false, errorx.Unknown
	}

	switch c.op {
	case isCompleted:
		if err != nil {
			return false, nil
		}

		if targetClaimedQuest.Status != entity.Accepted && targetClaimedQuest.Status != entity.AutoAccepted {
			return false, nil
		}

		return true, nil

	case isNotCompleted:
		if err != nil {
			return true, nil
		}

		if targetClaimedQuest.Status == entity.Rejected {
			return true, nil
		}

		return false, nil

	default:
		return false, errorx.New(errorx.BadRequest, "Invalid operator of Quest condition")
	}
}

// Data Condition
type dateConditionOpType string

var (
	dateBefore = enum.New(dateConditionOpType("before"))
	dateAfter  = enum.New(dateConditionOpType("after"))
)

type dateCondition struct {
	op   dateConditionOpType
	date time.Time
}

func newDateCondition(ctx router.Context, condition entity.Condition) (*dateCondition, error) {
	op, err := enum.ToEnum[dateConditionOpType](condition.Op)
	if err != nil {
		return nil, err
	}

	date, err := time.Parse("Jan 02 2006", condition.Value)
	if err != nil {
		return nil, err
	}

	return &dateCondition{op: op, date: date}, nil
}

func (c *dateCondition) Check(router.Context) (bool, error) {
	now := time.Now()

	switch c.op {
	case dateBefore:
		return now.Before(c.date), nil
	case dateAfter:
		return now.After(c.date), nil
	default:
		return false, errorx.New(errorx.BadRequest, "Invalid operator of Date condition")
	}
}
