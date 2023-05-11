package dateutil

import (
	"time"
)

// LastWeek returns the beginning of the lastweek of the current day.
func LastWeek(current time.Time) time.Time {
	beginningOfCurrentWeek := current.AddDate(0, 0, -int(current.Weekday()-time.Monday))
	lastWeek := beginningOfCurrentWeek.AddDate(0, 0, -6)
	return time.Date(lastWeek.Year(), lastWeek.Month(), lastWeek.Day(), 0, 0, 0, 0, lastWeek.Location())
}

// LastMonth returns the beginning of the last month of the current day.
func LastMonth(current time.Time) time.Time {
	beginningOfCurrentMonth := time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, current.Location())
	lastMonth := beginningOfCurrentMonth.AddDate(0, 0, -1)
	return time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, lastMonth.Location())
}
