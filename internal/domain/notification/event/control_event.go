package event

import "github.com/questx-lab/backend/internal/model"

type ReadyEvent struct {
	ChatMembers []model.ChatMember `json:"chat_members"`
	Communities []model.Community  `json:"communities"`
}

func (*ReadyEvent) Op() string {
	return "ready"
}
