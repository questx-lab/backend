package domain

import (
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode"

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

func checkCommunityHandle(handle string) error {
	if len(handle) < 4 {
		return errors.New("too short")
	}

	if len(handle) > 32 {
		return errors.New("too long")
	}

	ok, err := regexp.MatchString("^[a-z0-9_]*$", handle)
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("invalid name")
	}

	return nil
}

func checkCommunityDisplayName(displayName string) error {
	if len(displayName) < 4 {
		return errors.New("too short")
	}

	if len(displayName) > 32 {
		return errors.New("too long")
	}

	return nil
}

func generateCommunityHandle(displayName string) string {
	handle := []rune{}
	for _, c := range displayName {
		if isAsciiLetter(c) {
			handle = append(handle, unicode.ToLower(c))
		} else if c == ' ' {
			handle = append(handle, '_')
		}
	}

	return string(handle)
}

func isAsciiLetter(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') || c == '_'
}
