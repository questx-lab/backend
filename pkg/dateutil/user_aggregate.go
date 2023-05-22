package dateutil

import (
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/entity"
)

func GetCurrentValueByRange(r entity.UserAggregateRange) (string, error) {
	return GetValueByRange(time.Now(), r)
}

func GetPreviousValueByRange(r entity.UserAggregateRange) (string, error) {
	var t time.Time
	switch entity.UserAggregateRange(r) {
	case entity.UserAggregateRangeWeek:
		t = LastWeek(time.Now())
	case entity.UserAggregateRangeMonth:
		t = LastMonth(time.Now())
	case entity.UserAggregateRangeTotal:
		t = time.Now()
	}

	return GetValueByRange(t, r)
}

func GetValueByRange(t time.Time, r entity.UserAggregateRange) (string, error) {
	var val string
	switch entity.UserAggregateRange(r) {
	case entity.UserAggregateRangeWeek:
		year, week := t.ISOWeek()
		val = fmt.Sprintf("%d/%d", week, year)

	case entity.UserAggregateRangeMonth:
		val = fmt.Sprintf("%d/%d", t.Month(), t.Year())

	case entity.UserAggregateRangeTotal:
		val = "0/0"

	default:
		return "", fmt.Errorf("leader board range must be week, month, or total, but got %s", r)
	}

	return val, nil
}
