package entity

type Emoji struct {
	ID   string
	Name string
}

type ChatMessageReaction struct {
	MessageID int64
	UserID    string
	Emoji     Emoji
}
