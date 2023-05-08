package common

import (
	"errors"
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/domain/questclaim"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

const errReason = "Something is wrong"
const passReason = ""
const day = 24 * time.Hour
const week = 7 * day

func IsClaimable(
	ctx xcontext.Context,
	questFactory questclaim.Factory,
	claimedQuestRepo repository.ClaimedQuestRepository,
	quest entity.Quest,
) (reason string, err error) {
	// Check conditions.
	finalCondition := true
	if quest.ConditionOp == entity.Or && len(quest.Conditions) > 0 {
		finalCondition = false
	}

	var firstFailedCondition questclaim.Condition
	for _, c := range quest.Conditions {
		condition, err := questFactory.LoadCondition(ctx, c.Type, c.Data)
		if err != nil {
			return errReason, err
		}

		ok, err := condition.Check(ctx)
		if err != nil {
			return errReason, err
		}

		if firstFailedCondition == nil && !ok {
			firstFailedCondition = condition
		}

		if quest.ConditionOp == entity.And {
			finalCondition = finalCondition && ok
		} else {
			finalCondition = finalCondition || ok
		}
	}

	if !finalCondition {
		return firstFailedCondition.Statement(), nil
	}

	// Check recurrence.
	requestUserID := xcontext.GetRequestUserID(ctx)
	lastClaimedQuest, err := claimedQuestRepo.GetLastPendingOrAccepted(ctx, requestUserID, quest.ID)
	if err != nil {
		// The user has not claimed this quest yet.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return passReason, nil
		}

		return errReason, err
	}

	// If the user claimed the quest before, this quest cannot be claimed again until the next
	// recurrence.
	lastClaimedAt := lastClaimedQuest.CreatedAt
	switch quest.Recurrence {
	case entity.Once:
		return "This quest can only claim once", nil

	case entity.Daily:
		if lastClaimedAt.Day() != time.Now().Day() || time.Since(lastClaimedAt) > day {
			return passReason, nil
		}

		return "Please wait until the next day to claim this quest", nil

	case entity.Weekly:
		_, lastWeek := lastClaimedAt.ISOWeek()
		_, currentWeek := time.Now().ISOWeek()
		if lastWeek != currentWeek || time.Since(lastClaimedAt) > week {
			return passReason, nil
		}

		return "Please wait until the next week to claim this quest", nil

	case entity.Monthly:

		if lastClaimedAt.Month() != time.Now().Month() || lastClaimedAt.Year() != time.Now().Year() {
			return passReason, nil
		}

		return "Please wait until the next month to claim this quest", nil

	default:
		return errReason, fmt.Errorf("invalid recurrence %s", quest.Recurrence)
	}
}
