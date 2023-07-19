package model

type GetListMessageRequest struct {
	ChannelID     int64 `json:"channel_id"`
	LastMessageID int64 `json:"last_message_id"`
	Limit         int64 `json:"limit"`
}

type GetListMessageResponse struct {
	Messages []ChatMessage `json:"messages"`
}
