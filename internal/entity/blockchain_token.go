package entity

type BlockchainToken struct {
	Base

	Name       string
	Symbol     string
	Address    string     `gorm:"index:idx_blockchain_tokens_chain_address,unique"`
	Chain      string     `gorm:"index:idx_blockchain_tokens_chain_addresss,unique"`
	Blockchain Blockchain `gorm:"foreignKey:Chain;references:Name"`

	Decimals int
}
