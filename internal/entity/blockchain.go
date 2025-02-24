package entity

import "github.com/questx-lab/backend/pkg/enum"

type BlockchainConnectionType string

var (
	BlockchainConnectionRPC = enum.New(BlockchainConnectionType("rpc"))
)

type Blockchain struct {
	Name                 string `gorm:"primaryKey"`
	ID                   int64  `gorm:"unique"`
	DisplayName          string
	UseExternalRPC       bool
	UseEip1559           bool
	BlockTime            int
	AdjustTime           int
	ThresholdUpdateBlock int
	CurrencySymbol       string
	ExplorerURL          string
	XquestNFTAddress     string

	BlockchainConnections []BlockchainConnection `gorm:"foreignKey:Chain;references:Name"`
}

type BlockchainConnection struct {
	Chain string `gorm:"primaryKey"`
	URL   string `gorm:"primaryKey"`

	Type BlockchainConnectionType
}
