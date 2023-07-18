package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
)

type ChatChannelRepository interface {
	GetByID(ctx context.Context, id string) (*entity.ChatChannel, error)
	Create(context.Context, *entity.ChatChannel) error
}

type chatChannelRepository struct {
}

func NewChatChannelRepository() ChatChannelRepository {
	return &chatChannelRepository{}
}

func (r *chatChannelRepository) GetByID(ctx context.Context, id string) (*entity.ChatChannel, error) {
	return nil, nil
}
func (r *chatChannelRepository) Create(context.Context, *entity.ChatChannel) error {
	return nil
}
