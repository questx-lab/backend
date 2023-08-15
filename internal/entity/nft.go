package entity

import (
	"database/sql"
)

type NFTSet struct {
	Base
	CommunityID string
	Community   Community `gorm:"foreignKey:CommunityID"`
	Amount      int
	Title       string
	ImageUrl    string
	Chain       string
	Blockchain  Blockchain `gorm:"foreignKey:Chain;references:Name"`

	// extra fields
	ActiveCount  int
	PendingCount int
	ClaimedCount int
	FailureCount int
}

type NFT struct {
	ID BigInt

	SetID string
	Set   NFTSet `gorm:"foreignKey:SetID"`

	TransactionID sql.NullString
	Transaction   BlockchainTransaction `gorm:"foreignKey:TransactionID"`

	IsClaimed bool
}
