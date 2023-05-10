package dateutil

import (
	"time"
)

// IsYesterday returns true if target is yesterday of current, otherwise, return false.
func IsYesterday(target, current time.Time) bool {
	lastClaimYear, lastClaimMonth, lastClaimDay := target.Date()
	currentYear, currentMonth, currentDay := current.Date()
	return lastClaimYear == currentYear && lastClaimMonth == currentMonth && lastClaimDay+1 == currentDay
}

// Yesterday returns the beginning of the yesterday of the current day.
func Yesterday(current time.Time) time.Time {
	yesterday := current.AddDate(0, 0, -1)
	return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
}

// LastWeek returns the beginning of the lastweek of the current day.
func LastWeek(current time.Time) time.Time {
	beginningOfCurrentWeek := current.AddDate(0, 0, int(current.Weekday()-time.Monday))
	lastWeek := beginningOfCurrentWeek.AddDate(0, 0, -6)
	return time.Date(lastWeek.Year(), lastWeek.Month(), lastWeek.Day(), 0, 0, 0, 0, lastWeek.Location())
}

// LastMonth returns the beginning of the last month of the current day.
func LastMonth(current time.Time) time.Time {
	beginningOfCurrentMonth := time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, current.Location())
	lastMonth := beginningOfCurrentMonth.AddDate(0, 0, -1)
	return time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, lastMonth.Location())
}
