package entity

import "time"

type ChatChannel struct {
	ID          string
	Name        string
	CreatedBy   string
	CommunityID string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func (t *ChatChannel) TableName() string {
	return "channel"
}
