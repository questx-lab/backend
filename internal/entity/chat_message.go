package entity

import "time"

type ChatMessage struct {
	ID        int64
	UserID    string
	ChannelID int64
	Bucket    int64
	ReplyTo   int64
	Message   string
	CreatedAt time.Time
}

func (t *ChatMessage) TableName() string {
	return "messages"
}
