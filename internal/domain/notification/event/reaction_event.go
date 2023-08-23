package event

import (
	"encoding/json"

	"github.com/questx-lab/backend/internal/entity"
)

// REACTION ADDED EVENT
type ReactionAddedEvent struct {
	ChannelID int64        `json:"channel_id"`
	MessageID int64        `json:"message_id"`
	UserID    string       `json:"user_id"`
	Emoji     entity.Emoji `json:"emoji"`
}

func (ReactionAddedEvent) Op() string {
	return "reaction_added"
}

func (e *ReactionAddedEvent) Unmarshal(b []byte) error {
	return json.Unmarshal(b, e)
}

// REACTION REMOVED EVENT
type ReactionRemovedEvent struct {
	ChannelID int64        `json:"channel_id"`
	MessageID int64        `json:"message_id"`
	UserID    string       `json:"user_id"`
	Emoji     entity.Emoji `json:"emoji"`
}

func (ReactionRemovedEvent) Op() string {
	return "reaction_removed"
}

func (e *ReactionRemovedEvent) Unmarshal(b []byte) error {
	return json.Unmarshal(b, e)
}
