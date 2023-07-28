package model

import "github.com/questx-lab/backend/internal/entity"

type GetMessagesRequest struct {
	ChannelID int64 `json:"channel_id"`
	Before    int64 `json:"before"`
	Limit     int64 `json:"limit"`
}

type GetMessagesResponse struct {
	Messages []ChatMessage `json:"messages"`
}

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

type RemoveReactionRequest struct {
	ChannelID int64        `json:"channel_id"`
	MessageID int64        `json:"message_id"`
	Emoji     entity.Emoji `json:"emoji"`
}

type RemoveReactionResponse struct{}

type GetUserReactionsRequest struct {
	ChannelID int64        `json:"channel_id"`
	MessageID int64        `json:"message_id"`
	Emoji     entity.Emoji `json:"emoji"`
	Limit     int64
}

type GetUserReactionsResponse struct {
	Users []User `json:"users"`
}

type DeleteMessageRequest struct {
	ChannelID int64 `json:"channel_id"`
	MessageID int64 `json:"message_id"`
}

type DeleteMessageResponse struct{}

type GetChannelsRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GetChannelsResponse struct {
	Channels []ChatChannel `json:"channels"`
}
