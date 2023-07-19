package entity

type ChatMember struct {
	UserID string
	User   User `gorm:"foreignKey:UserID"`

	ChannelID int64
	Channel   ChatChannel `gorm:"foreignKey:ChannelID"`

	LastReadMessageID int64
}
