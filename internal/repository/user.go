package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository interface {
	Create(ctx context.Context, data *entity.User) error
	UpdateByID(ctx context.Context, id string, data *entity.User) error
	RetrieveByID(ctx context.Context, id string) (*entity.User, error)
	RetrieveByAddress(ctx context.Context, address string) (*entity.User, error)
	RetrieveByServiceID(
		ctx context.Context, service, serviceUserID string) (*entity.User, error)
	DeleteByID(ctx context.Context, id string) error
	UpsertByID(ctx context.Context, id string, data *entity.User) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, data *entity.User) error {
	return r.db.Create(data).Error
}

func (r *userRepository) UpdateByID(ctx context.Context, id string, data *entity.User) error {
	panic("not implemented") // TODO: Implement
}

func (r *userRepository) RetrieveByID(ctx context.Context, id string) (*entity.User, error) {
	var record entity.User
	if err := r.db.Where("id=?", id).Take(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *userRepository) RetrieveByAddress(ctx context.Context, address string) (*entity.User, error) {
	var record entity.User
	if err := r.db.Where("address=?", address).Take(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *userRepository) RetrieveByServiceID(
	ctx context.Context, service, serviceUserID string,
) (*entity.User, error) {
	var record entity.User
	err := r.db.
		Model(&entity.User{}).
		Where("oauth2.service=? AND oauth2.service_user_id=?", service, serviceUserID).
		Joins("join oauth2 on users.id=oauth2.user_id").
		Take(&record).Error
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (r *userRepository) DeleteByID(ctx context.Context, id string) error {
	panic("not implemented") // TODO: Implement
}

func (r *userRepository) UpsertByID(ctx context.Context, id string, data *entity.User) error {
	var record entity.User
	err := r.db.
		Model(&record).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoNothing: true,
		}).Create(&data).Error
	if err != nil {
		return err
	}

	return nil
}
