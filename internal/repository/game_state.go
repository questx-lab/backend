package repository

import (
	"context"
)

type RoomRepository interface {
	GetByRoomID(context.Context, string) error
}

type roomRepository struct {
}

func NewRoomRepository() RoomRepository {
	return &roomRepository{}
}

func (r *roomRepository) GetByRoomID(ctx context.Context, roomID string) error {
	panic("not implemented")
}
