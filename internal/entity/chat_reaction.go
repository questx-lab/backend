package entity

import "github.com/scylladb/gocqlx/v2"

type Emoji struct {
	gocqlx.UDT `json:"-"`
	Name       string `json:"name"`
}

type ChatReaction struct {
	MessageID int64
	Emoji     Emoji
	UserIds   []string
}

func (ChatReaction) TableName() string {
	return "chat_reactions"
}
