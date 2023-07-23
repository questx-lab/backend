package entity

import "database/sql"

type PayReward struct {
	Base

	TokenID         string
	BlockchainToken BlockchainToken `gorm:"foreignKey:TokenID"`

	TransactionID sql.NullString
	Transaction   BlockchainTransaction `gorm:"foreignKey:TransactionID"`

	FromCommunityID sql.NullString
	FromCommunity   Community `gorm:"foreignKey:FromCommunityID"`

	ToUserID string
	ToUser   User `gorm:"foreignKey:ToUserID"`

	ToAddress string
	Amount    float64

	// Reason of pay reward.
	ClaimedQuestID sql.NullString
	ClaimedQuest   ClaimedQuest `gorm:"foreignKey:ClaimedQuestID"`

	ReferralCommunityID sql.NullString
	ReferralCommunity   Community `gorm:"foreignKey:ReferralCommunityID"`

	LotteryWinnerID sql.NullString
	LotteryWinner   LotteryWinner `gorm:"foreignKey:LotteryWinnerID"`
}
