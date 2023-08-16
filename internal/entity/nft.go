package entity

import (
	"database/sql"
)

type NFTSet struct {
	ID            BigInt
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
	ID BigInt

	SetID string
	Set   NFTSet `gorm:"foreignKey:SetID"`

	TransactionID sql.NullString
	Transaction   BlockchainTransaction `gorm:"foreignKey:TransactionID"`
}
