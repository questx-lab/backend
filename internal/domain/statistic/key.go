package statistic

import (
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
)

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
