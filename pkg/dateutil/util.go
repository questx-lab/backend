package dateutil

import (
	"time"
)

func Date(current time.Time) time.Time {
	return time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, current.Location())
}

func IsYesterday(target, current time.Time) bool {
	lastClaimYear, lastClaimMonth, lastClaimDay := target.Date()
	currentYear, currentMonth, currentDay := current.Date()
	return lastClaimYear == currentYear && lastClaimMonth == currentMonth && lastClaimDay+1 == currentDay
}

func IsToday(target, current time.Time) bool {
	lastClaimYear, lastClaimMonth, lastClaimDay := target.Date()
	currentYear, currentMonth, currentDay := current.Date()
	return lastClaimYear == currentYear && lastClaimMonth == currentMonth && lastClaimDay == currentDay
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

// NextMonth returns the beginning of the next month of the current day.
func NextMonth(current time.Time) time.Time {
	currentYear := current.Year()
	currentMonth := current.Month()
	if currentMonth == time.December {
		currentYear += 1
		currentMonth = time.January
	} else {
		currentMonth += 1
	}

	return time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, current.Location())
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

// CurrentWeek returns the beginning of the current week.
func CurrentWeek(current time.Time) time.Time {
	weekday := current.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	currentWeek := current.AddDate(0, 0, -int(weekday-time.Monday))
	return time.Date(currentWeek.Year(), currentWeek.Month(), currentWeek.Day(), 0, 0, 0, 0, currentWeek.Location())
}

// NextDay returns the beginning of the next day of the current day.
func NextDay(current time.Time) time.Time {
	nextDay := current.AddDate(0, 0, 1)
	return time.Date(nextDay.Year(), nextDay.Month(), nextDay.Day(), 0, 0, 0, 0, nextDay.Location())
}
