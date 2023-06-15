package entity

type PayReward struct {
	Base

	ToUserID string
	ToUser   User `gorm:"foreignKey:ToUserID"`

	// Note contains the reason of this transaction in case of not come from a
	// claimed quest.
	Note    string
	Address string
	Token   string
	Amount  float64

	IsReceived bool
}
