package event

import "github.com/questx-lab/backend/internal/model"

type ReadyEvent struct {
	User        model.User        `json:"user"`
	Communities []model.Community `json:"communities"`
}

func (*ReadyEvent) Op() string {
	return "ready"
}
