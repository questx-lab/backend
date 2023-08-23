package event

import (
	"encoding/json"

	"github.com/questx-lab/backend/internal/model"
)

// CHANNEL CREATED EVENT
type ChannelCreatedEvent struct {
	model.ChatChannel
}

func (ChannelCreatedEvent) Op() string {
	return "channel_created"
}

func (e *ChannelCreatedEvent) Unmarshal(b []byte) error {
	return json.Unmarshal(b, e)
}

// CHANNEL UPDATED EVENT
type ChannelUpdatedEvent struct {
	model.ChatChannel
}

func (ChannelUpdatedEvent) Op() string {
	return "channel_updated"
}

func (e *ChannelUpdatedEvent) Unmarshal(b []byte) error {
	return json.Unmarshal(b, e)
}

// CHANNEL DELETED EVENT
type ChannelDeletedEvent struct {
	CommunityID string `json:"community_id"`
	ChannelID   int64  `json:"channel_id"`
}

func (ChannelDeletedEvent) Op() string {
	return "channel_deleted"
}
func (e *ChannelDeletedEvent) Unmarshal(b []byte) error {
	return json.Unmarshal(b, e)
}
