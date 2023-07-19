package entity

type ChatMessageReactionStatistic struct {
	MessageID  int64
	ReactionID int64
	Quantity   uint64
}

func (t *ChatMessageReactionStatistic) TableName() string {
	return "message_reaction_statistics"
}
