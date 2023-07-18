package entity

import "time"

type ChatMessage struct {
	ID         int64
	ChannelID  int64
	UserID     string
	ReplyTo    int64
	Content    string
	Attachment []string
	CreatedAt  time.Time
}

func (t *ChatMessage) TableName() string {
	return "messages"
}
