package model

type GetListMessageRequest struct {
	ChannelID     string `json:"channel_id"`
	LastMessageID string `json:"last_message_id"`
	Limit         int64  `json:"limit"`
}

type GetListMessageResponse struct {
	Messages []ChatMessage `json:"messages"`
}

type ChatMessage struct {
	MessageID string         `json:"message_id"`
	Message   string         `json:"message"`
	Reactions []ChatReaction `json:"reactions"`
}

type ChatReaction struct {
	ReactionID string `json:"reaction_id"`
	Quantity   int64  `json:"quantity"`
}
