package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
)

type ChatChannelRepository interface {
	GetByID(ctx context.Context, id string) (*entity.ChatChannel, error)
	Create(ctx context.Context, *entity.ChatChannel)  error
}

type chatChannelRepository struct {
}

func NewChatChannelRepository() ChatChannelRepository {
	return &chatChannelRepository{}
}
