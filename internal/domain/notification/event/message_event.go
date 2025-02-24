package event

import "github.com/questx-lab/backend/internal/model"

// MESSAGE CREATED EVENT
type MessageCreatedEvent struct {
	model.ChatMessage
}

func (*MessageCreatedEvent) Op() string {
	return "message_created"
}

// MESSAGE UPDATED EVENT
type MessageUpdatedEvent struct {
	model.ChatMessage
}

func (*MessageUpdatedEvent) Op() string {
	return "messge_updated"
}

// MESSAGE UPDATED EVENT
type MessageDeletedEvent struct {
	MessageID int64 `json:"message_id"`
}

func (*MessageDeletedEvent) Op() string {
	return "message_deleted"
}
