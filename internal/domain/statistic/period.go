package statistic

import (
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/dateutil"
)

func ToPeriodWithTime(periodString string, current time.Time) (entity.LeaderBoardPeriodType, error) {
	switch periodString {
	case "week":
		return entity.NewLeaderBoardPeriodWeek(current), nil
	case "month":
		return entity.NewLeaderBoardPeriodMonth(current), nil
	}

	return nil, fmt.Errorf("invalid period, expected week or month, but got %s", periodString)
}

func ToPeriod(periodString string) (entity.LeaderBoardPeriodType, error) {
	switch periodString {
	case "week":
		return entity.NewLeaderBoardPeriodWeek(time.Now()), nil
	case "month":
		return entity.NewLeaderBoardPeriodMonth(time.Now()), nil
	}

	return nil, fmt.Errorf("invalid period, expected week or month, but got %s", periodString)
}

func ToLastPeriod(periodString string) (entity.LeaderBoardPeriodType, error) {
	switch periodString {
	case "week":
		return entity.NewLeaderBoardPeriodWeek(dateutil.LastWeek(time.Now())), nil
	case "month":
		return entity.NewLeaderBoardPeriodMonth(dateutil.LastMonth(time.Now())), nil
	}

	return nil, fmt.Errorf("invalid period, expected week or month, but got %s", periodString)
}
