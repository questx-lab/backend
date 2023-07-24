package event

import "github.com/questx-lab/backend/internal/model"

// CHANNEL CREATED EVENT
type ChannelCreatedEvent struct {
	model.ChatChannel
}

func (*ChannelCreatedEvent) Op() string {
	return "channel_created"
}

// CHANNEL UPDATED EVENT
type ChannelUpdatedEvent struct {
	model.ChatChannel
}

func (*ChannelUpdatedEvent) Op() string {
	return "channel_updated"
}

// CHANNEL DELETED EVENT
type ChannelDeletedEvent struct {
	ChannelID int64 `json:"channel_id"`
}

func (*ChannelDeletedEvent) Op() string {
	return "channel_deleted"
}
