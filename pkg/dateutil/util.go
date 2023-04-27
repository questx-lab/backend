package dateutil

import (
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/entity"
)

func GetCurrentValueByRange(ra entity.UserAggregateRange) (string, error) {
	now := time.Now()

	var val string

	switch entity.UserAggregateRange(ra) {
	case entity.UserAggregateRangeWeek:
		year, week := now.ISOWeek()
		val = fmt.Sprintf(`week/%d/%d`, week, year)
	case entity.UserAggregateRangeMonth:
		month := now.Month()
		year := now.Year()
		val = fmt.Sprintf(`month/%d/%d`, month, year)
	case entity.UserAggregateRangeTotal:
	default:
		return "", fmt.Errorf("leader board range must be week, month, total. but got %s", ra)
	}
	return val, nil
}
