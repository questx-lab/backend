package entity

type BlockchainToken struct {
	Base

	Name       string
	Token      string     `gorm:"index:idx_blockchain_tokens_chain_token,unique"`
	Chain      string     `gorm:"index:idx_blockchain_tokens_chain_token,unique"`
	Blockchain Blockchain `gorm:"foreignKey:Chain;references:Name"`

	Address  string
	Decimals int
}
