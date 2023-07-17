package entity

type ChatMessage struct {
	Base
	UserID    string
	ChannelID string
	ReplyTo   string
	Message   string
}

func (t *ChatMessage) TableName() string {
	return "messages"
}
