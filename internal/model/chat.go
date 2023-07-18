package model

import "github.com/questx-lab/backend/internal/entity"

type CreateMessageRequest struct {
	ChannelID   int64               `json:"channel_id"`
	Content     string              `json:"content"`
	Attachments []entity.Attachment `json:"attachments,omitempty"`
}

type CreateMessageResponse struct {
	ID int64 `json:"id"`
}
