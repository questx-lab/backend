package entity

import (
	"fmt"
	"time"

	"github.com/questx-lab/backend/pkg/dateutil"
)

type LeaderBoardPeriodType interface {
	Period() string
	Start() time.Time
	End() time.Time
}

type LeaderBoardPeriodWeek struct {
	current time.Time
}

func NewLeaderBoardPeriodWeek(current time.Time) LeaderBoardPeriodWeek {
	return LeaderBoardPeriodWeek{current: current}
}

func (p LeaderBoardPeriodWeek) Period() string {
	year, week := p.current.ISOWeek()
	return fmt.Sprintf("%d:%d", week, year)
}

func (p LeaderBoardPeriodWeek) Start() time.Time {
	return dateutil.CurrentWeek(p.current)
}

func (p LeaderBoardPeriodWeek) End() time.Time {
	return p.Start().AddDate(0, 0, 7)
}

type LeaderBoardPeriodMonth struct {
	current time.Time
}

func NewLeaderBoardPeriodMonth(current time.Time) LeaderBoardPeriodMonth {
	return LeaderBoardPeriodMonth{current: current}
}

func (p LeaderBoardPeriodMonth) Period() string {
	return fmt.Sprintf("%s:%d", p.current.Month(), p.current.Year())
}

func (p LeaderBoardPeriodMonth) Start() time.Time {
	return time.Date(p.current.Year(), p.current.Month(), 1, 0, 0, 0, 0, p.current.Location())
}

func (p LeaderBoardPeriodMonth) End() time.Time {
	return p.Start().AddDate(0, 1, 0)
}

type UserStatistic struct {
	UserID      string
	CommunityID string
	Points      uint64
	Quests      uint64
}
