package dateutil

import (
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/entity"
)

func GetCurrentValueByRange(ra entity.AchievementRange) (string, error) {
	now := time.Now()

	var val string

	switch entity.AchievementRange(ra) {
	case entity.AchievementRangeWeek:
		year, week := now.ISOWeek()
		val = fmt.Sprintf(`week/%d/%d`, week, year)
	case entity.AchievementRangeMonth:
		month := now.Month()
		year := now.Year()
		val = fmt.Sprintf(`month/%d/%d`, month, year)
	case entity.AchievementRangeTotal:
	default:
		return "", fmt.Errorf("Leader board range must be week, month, total")
	}
	return val, nil
}
