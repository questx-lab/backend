package event

import (
	"github.com/questx-lab/backend/internal/entity"
)

// REACTION ADDED EVENT
type ReactionAddedEvent struct {
	MessageID int64        `json:"message_id"`
	UserID    string       `json:"user_id"`
	Emoji     entity.Emoji `json:"emoji"`
}

func (*ReactionAddedEvent) Op() string {
	return "reaction_added"
}

// REACTION REMOVED EVENT
type ReactionRemovedEvent struct {
	MessageID int64        `json:"message_id"`
	UserID    string       `json:"user_id"`
	Emoji     entity.Emoji `json:"emoji"`
}

func (*ReactionRemovedEvent) Op() string {
	return "reaction_removed"
}
