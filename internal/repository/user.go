package repository

import (
	"context"
	"database/sql"

	"github.com/questx-lab/backend/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, data *entity.User) error
	UpdateByID(ctx context.Context, id string, data *entity.User) error
	RetrieveByID(ctx context.Context, id string) (*entity.User, error)
	DeleteByID(ctx context.Context, id string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, data *entity.User) error {
	panic("not implemented") // TODO: Implement
}

func (r *userRepository) UpdateByID(ctx context.Context, id string, data *entity.User) error {
	panic("not implemented") // TODO: Implement
}

func (r *userRepository) RetrieveByID(ctx context.Context, id string) (*entity.User, error) {
	panic("not implemented") // TODO: Implement
}

func (r *userRepository) DeleteByID(ctx context.Context, id string) error {
	panic("not implemented") // TODO: Implement
}
