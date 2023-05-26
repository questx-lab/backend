package domain

import (
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/dateutil"
)

func rewardEntityToModel(entityRewards []entity.Reward) []model.Reward {
	modelRewards := []model.Reward{}
	for _, r := range entityRewards {
		modelRewards = append(modelRewards, model.Reward{Type: string(r.Type), Data: r.Data})
	}
	return modelRewards
}

func conditionEntityToModel(entityConditions []entity.Condition) []model.Condition {
	modelConditions := []model.Condition{}
	for _, r := range entityConditions {
		modelConditions = append(modelConditions, model.Condition{Type: string(r.Type), Data: r.Data})
	}
	return modelConditions
}

func stringToPeriodWithTime(periodString string, current time.Time) (entity.LeaderBoardPeriodType, error) {
	switch periodString {
	case "week":
		return entity.NewLeaderBoardPeriodWeek(current), nil
	case "month":
		return entity.NewLeaderBoardPeriodMonth(current), nil
	}

	return nil, fmt.Errorf("invalid period, expected week or month, but got %s", periodString)
}

func stringToPeriod(periodString string) (entity.LeaderBoardPeriodType, error) {
	switch periodString {
	case "week":
		return entity.NewLeaderBoardPeriodWeek(time.Now()), nil
	case "month":
		return entity.NewLeaderBoardPeriodMonth(time.Now()), nil
	}

	return nil, fmt.Errorf("invalid period, expected week or month, but got %s", periodString)
}

func stringToLastPeriod(periodString string) (entity.LeaderBoardPeriodType, error) {
	switch periodString {
	case "week":
		return entity.NewLeaderBoardPeriodWeek(dateutil.LastWeek(time.Now())), nil
	case "month":
		return entity.NewLeaderBoardPeriodMonth(dateutil.LastMonth(time.Now())), nil
	}

	return nil, fmt.Errorf("invalid period, expected week or month, but got %s", periodString)
}

func redisKeyPointLeaderBoard(communityID string, period entity.LeaderBoardPeriodType) string {
	return fmt.Sprintf("%s:point:%s", communityID, period.Period())
}

func redisKeyQuestLeaderBoard(communityID string, period entity.LeaderBoardPeriodType) string {
	return fmt.Sprintf("%s:quest:%s", communityID, period.Period())
}

func redisKeyLeaderBoard(orderedBy, communityID string, period entity.LeaderBoardPeriodType) (string, error) {
	switch orderedBy {
	case "point":
		return redisKeyPointLeaderBoard(communityID, period), nil
	case "quest":
		return redisKeyQuestLeaderBoard(communityID, period), nil
	}

	return "", fmt.Errorf("expected ordered by point or quest, but got %s", orderedBy)
}
