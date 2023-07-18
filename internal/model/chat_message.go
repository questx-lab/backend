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
	MessageID string
	Message   string
	Reaction  map[string]int
}
