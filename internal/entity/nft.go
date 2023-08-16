package entity

import (
	"database/sql"
)

type NFTSet struct {
	SnowFlakeBase
	CommunityID   string
	Community     Community `gorm:"foreignKey:CommunityID"`
	Title         string
	Description   string
	ImageUrl      string
	Chain         string
	NFTAddress    string
	Blockchain    Blockchain `gorm:"foreignKey:Chain;references:Name"`
	CreatedBy     string
	CreatedByUser User `gorm:"foreignKey:CreatedBy"`
}

type NFT struct {
	SnowFlakeBase

	SetID int64
	Set   NFTSet `gorm:"foreignKey:SetID"`

	TransactionID sql.NullString
	Transaction   BlockchainTransaction `gorm:"foreignKey:TransactionID"`
}
