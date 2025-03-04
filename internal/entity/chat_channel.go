package entity

type ChatChannel struct {
	SnowFlakeBase

	CommunityID string
	Community   Community `gorm:"foreignKey:CommunityID"`

	Name          string
	LastMessageID int64
	Description   string
}
