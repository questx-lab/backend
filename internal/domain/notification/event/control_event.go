package event

import "github.com/questx-lab/backend/internal/model"

type ReadyEvent struct {
	ChatMembers []model.ChatMember `json:"chat_members"`
}

func (*ReadyEvent) Op() string {
	return "ready"
}
