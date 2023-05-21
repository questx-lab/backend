package dateutil

import (
	"time"
)

func IsYesterday(target, current time.Time) bool {
	lastClaimYear, lastClaimMonth, lastClaimDay := target.Date()
	currentYear, currentMonth, currentDay := current.Date()
	return lastClaimYear == currentYear && lastClaimMonth == currentMonth && lastClaimDay+1 == currentDay
}

// LastWeek returns the beginning of the lastweek of the current day.
func LastWeek(current time.Time) time.Time {
	weekday := current.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	beginningOfCurrentWeek := current.AddDate(0, 0, -int(weekday-time.Monday))
	lastWeek := beginningOfCurrentWeek.AddDate(0, 0, -6)
	return time.Date(lastWeek.Year(), lastWeek.Month(), lastWeek.Day(), 0, 0, 0, 0, lastWeek.Location())
}

// LastMonth returns the beginning of the last month of the current day.
func LastMonth(current time.Time) time.Time {
	beginningOfCurrentMonth := time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, current.Location())
	lastMonth := beginningOfCurrentMonth.AddDate(0, 0, -1)
	return time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, lastMonth.Location())
}

// NextWeek returns the beginning of the next week of the current day.
func NextWeek(current time.Time) time.Time {
	weekday := current.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	nextWeek := current.AddDate(0, 0, int(7-weekday+1))
	return time.Date(nextWeek.Year(), nextWeek.Month(), nextWeek.Day(), 0, 0, 0, 0, nextWeek.Location())
}
