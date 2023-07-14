package entity

type ChatMessage struct {
	ID string
}

func (t *ChatMessage) TableName() string {
	return "messages"
}
