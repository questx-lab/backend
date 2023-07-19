package model

type GetListMessageRequest struct {
	ChannelID     int64 `json:"channel_id"`
	LastMessageID int64 `json:"last_message_id"`
	Limit         int64 `json:"limit"`
}

type GetListMessageResponse struct {
	Messages []ChatMessage `json:"messages"`
}

type ChatMessage struct {
	MessageID int64          `json:"message_id"`
	Message   string         `json:"message"`
	Reactions []ChatReaction `json:"reactions"`
}

type ChatReaction struct {
	ReactionID int64 `json:"reaction_id"`
	Quantity   int64 `json:"quantity"`
}
