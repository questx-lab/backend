package model

import "github.com/questx-lab/backend/internal/entity"

type CreateChannelRequest struct {
	CommunityHandle string `json:"community_handle"`
	ChannelName     string `json:"channel_name"`
}

type CreateChannelResponse struct {
	ID int64 `json:"id"`
}

type CreateMessageRequest struct {
	ChannelID   int64               `json:"channel_id"`
	Content     string              `json:"content"`
	Attachments []entity.Attachment `json:"attachments,omitempty"`
}

type CreateMessageResponse struct {
	ID int64 `json:"id"`
}

type AddReactionRequest struct {
	ChannelID int64        `json:"channel_id"`
	MessageID int64        `json:"message_id"`
	Emoji     entity.Emoji `json:"emoji"`
}

type AddReactionResponse struct{}
