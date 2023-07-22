package entity

import "time"

type LotteryEvent struct {
	Base

	CommunityID string
	Community   Community `gorm:"foreignKey:CommunityID"`

	StartTime      time.Time
	EndTime        time.Time
	MaxTickets     int
	UsedTickets    int
	PointPerTicket uint64
}

type LotteryPrize struct {
	Base

	LotteryEventID string
	LotteryEvent   LotteryEvent `gorm:"foreignKey:LotteryEventID"`

	Points  int
	Rewards Array[Reward]

	AvailableRewards int
	WonRewards       int
}

type LotteryWinner struct {
	Base

	LotteryPrizeID string
	LotteryPrize   LotteryPrize `gorm:"foreignKey:LotteryPrizeID"`

	UserID string
	User   User `gorm:"foreignKey:UserID"`

	IsClaimed bool
}
