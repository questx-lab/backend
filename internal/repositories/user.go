package repositories

import (
	"context"

	"sisu-network/gateway/internal/entities"
)

type UserRepository interface {
	Create(ctx context.Context, data *entities.User) error
	UpdateByID(ctx context.Context, id string, data *entities.User) error
	RetrieveByID(ctx context.Context, id string) (*entities.User, error)
	DeleteByID(ctx context.Context, id string) error
}
