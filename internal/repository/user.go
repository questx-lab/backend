package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, data *entity.User) error
	UpdateByID(ctx context.Context, id string, data *entity.User) error
	RetrieveByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	RetrieveByAddress(ctx context.Context, address string) (*entity.User, error)
	RetrieveByServiceID(
		ctx context.Context, service, serviceUserID string) (*entity.User, error)
	DeleteByID(ctx context.Context, id string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, data *entity.User) error {
	return nil
}

func (r *userRepository) UpdateByID(ctx context.Context, id string, data *entity.User) error {
	panic("not implemented") // TODO: Implement
}

func (r *userRepository) RetrieveByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	panic("not implemented") // TODO: Implement
}

func (r *userRepository) RetrieveByAddress(ctx context.Context, address string) (*entity.User, error) {
	panic("not implemented") // TODO: Implement
}

func (r *userRepository) RetrieveByServiceID(
	ctx context.Context, serviceID, serviceUserID string,
) (*entity.User, error) {
	return nil, errors.New("not found user")
}

func (r *userRepository) DeleteByID(ctx context.Context, id string) error {
	panic("not implemented") // TODO: Implement
}
