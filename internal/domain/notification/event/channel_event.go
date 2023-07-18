package event

import "github.com/questx-lab/backend/internal/model"

// CHANNEL CREATED EVENT
type ChannelCreatedEvent model.Channel

func (*ChannelCreatedEvent) Op() string {
	return "channel_created"
}

// CHANNEL UPDATED EVENT
type ChannelUpdatedEvent model.Channel

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
