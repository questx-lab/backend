package entity

import "github.com/scylladb/gocqlx/v2"

type Attachment struct {
	gocqlx.UDT `json:"-"`
	URL        string `json:"url"`
}

type ChatMessage struct {
	ID          int64
	Bucket      int64
	ChannelID   int64
	AuthorID    string
	ReplyTo     int64
	Content     string
	Attachments []Attachment
}

func (t *ChatMessage) TableName() string {
	return "chat_messages"
}
