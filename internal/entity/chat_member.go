package entity

type ChatMember struct {
	UserID string
	User   User `gorm:"foreignKey:UserID"`

	ChannelID string
	Channel   ChatChannel `gorm:"foreignKey:ChatID"`

	LastReadMessageID string
}
