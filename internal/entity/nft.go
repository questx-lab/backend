package entity

import (
	"time"
)

type NonFungibleToken struct {
	SnowFlakeBase

	CommunityID string
	Community   Community `gorm:"foreignKey:CommunityID"`

	CreatedBy     string
	CreatedByUser User `gorm:"foreignKey:CreatedBy"`

	Chain      string
	Blockchain Blockchain `gorm:"foreignKey:Chain;references:Name"`

	Title       string
	Description string
	ImageUrl    string

	NumberOfClaimed int
}

type NonFungibleTokenMintHistory struct {
	NonFungibleTokenID int64
	NonFungibleToken   NonFungibleToken `gorm:"foreignKey:NonFungibleTokenID"`

	CreatedAt time.Time

	TransactionID string
	Transaction   BlockchainTransaction `gorm:"foreignKey:TransactionID"`

	Amount int
}
