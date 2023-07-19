package entity

type ChatReactionStatistic struct {
	MessageID int64
	Emoji     Emoji
	Count     int64
}

func (ChatReactionStatistic) TableName() string {
	return "chat_reaction_statistics"
}
