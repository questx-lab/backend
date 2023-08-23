package event

import (
	"encoding/json"

	"github.com/questx-lab/backend/internal/model"
)

// MESSAGE CREATED EVENT
type MessageCreatedEvent struct {
	model.ChatMessage
}

func (MessageCreatedEvent) Op() string {
	return "message_created"
}

func (e *MessageCreatedEvent) Unmarshal(b []byte) error {
	return json.Unmarshal(b, e)
}

// MESSAGE UPDATED EVENT
type MessageUpdatedEvent struct {
	model.ChatMessage
}

func (MessageUpdatedEvent) Op() string {
	return "messge_updated"
}

func (e *MessageUpdatedEvent) Unmarshal(b []byte) error {
	return json.Unmarshal(b, e)
}

// MESSAGE UPDATED EVENT
type MessageDeletedEvent struct {
	MessageID int64 `json:"message_id"`
}

func (MessageDeletedEvent) Op() string {
	return "message_deleted"
}

func (e *MessageDeletedEvent) Unmarshal(b []byte) error {
	return json.Unmarshal(b, e)
}
