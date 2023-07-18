package entity

import "time"

type ChatMessage struct {
	ID        string
	UserID    string
	ChannelID string
	Bucket    int64
	ReplyTo   string
	Message   string
	CreatedAt time.Time
}

func (t *ChatMessage) TableName() string {
	return "messages"
}
