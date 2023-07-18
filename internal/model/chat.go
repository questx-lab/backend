package model

type CreateMessageRequest struct {
	ChannelID   int64        `json:"channel_id"`
	Content     string       `json:"content"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type CreateMessageResponse struct {
	ID string `json:"id"`
}
