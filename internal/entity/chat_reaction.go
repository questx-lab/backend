package entity

import "github.com/scylladb/gocqlx/v2"

type Emoji struct {
	gocqlx.UDT
	Name string `json:"name"`
}

type ChatReaction struct {
	MessageID int64
	UserID    string
	Emoji     Emoji
}

func (ChatReaction) TableName() string {
	return "chat_reactions"
}
