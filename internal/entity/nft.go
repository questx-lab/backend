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

	Title    string
	ImageUrl string
}

type NonFungibleTokenMintHistory struct {
	NonFungibleTokenID int64            `gorm:"primaryKey"`
	NonFungibleToken   NonFungibleToken `gorm:"foreignKey:NonFungibleTokenID"`

	CreatedAt time.Time

	TransactionID string
	Transaction   BlockchainTransaction `gorm:"foreignKey:TransactionID"`

	Count int
}
