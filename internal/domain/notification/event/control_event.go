package event

import (
	"encoding/json"

	"github.com/questx-lab/backend/internal/model"
)

type ReadyEvent struct {
	ChatMembers []model.ChatMember `json:"chat_members"`
	Communities []model.Community  `json:"communities"`
}

func (ReadyEvent) Op() string {
	return "ready"
}

func (e *ReadyEvent) Unmarshal(b []byte) error {
	return json.Unmarshal(b, e)
}
