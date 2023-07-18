package entity

type ChatMessageReactionStatistic struct {
	MessageID  string
	ReactionID string
	Quantity   uint64
}

func (t *ChatMessageReactionStatistic) TableName() string {
	return "message_reaction_statistics"
}
