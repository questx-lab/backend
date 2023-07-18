package entity

import "time"

type Attachment struct {
	URL string `json:"url"`
}

type ChatMessage struct {
	ID          int64
	ChannelID   int64
	AuthorID    string
	ReplyTo     int64
	Content     string
	Attachments []Attachment
	CreatedAt   time.Time
}

func (t *ChatMessage) TableName() string {
	return "messages"
}
