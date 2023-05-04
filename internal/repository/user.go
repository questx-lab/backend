package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm/clause"
)

type UserRepository interface {
	Create(ctx xcontext.Context, data *entity.User) error
	UpdateByID(ctx xcontext.Context, id string, data *entity.User) error
	GetByID(ctx xcontext.Context, id string) (*entity.User, error)
	GetByIDs(ctx xcontext.Context, ids []string) ([]entity.User, error)
	GetByAddress(ctx xcontext.Context, address string) (*entity.User, error)
	GetByServiceUserID(ctx xcontext.Context, service, serviceUserID string) (*entity.User, error)
	DeleteByID(ctx xcontext.Context, id string) error
	UpsertByID(ctx xcontext.Context, id string, data *entity.User) error
}

type userRepository struct {
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) Create(ctx xcontext.Context, data *entity.User) error {
	return ctx.DB().Create(data).Error
}

func (r *userRepository) UpdateByID(ctx xcontext.Context, id string, data *entity.User) error {
	panic("not implemented") // TODO: Implement
}

func (r *userRepository) GetByID(ctx xcontext.Context, id string) (*entity.User, error) {
	var record entity.User
	if err := ctx.DB().Where("id=?", id).Take(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

func (r *userRepository) GetByIDs(ctx xcontext.Context, ids []string) ([]entity.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var record []entity.User
	if err := ctx.DB().Where("id IN (?)", ids).Find(&record).Error; err != nil {
		return nil, err
	}

	return record, nil
}

func (r *userRepository) GetByAddress(ctx xcontext.Context, address string) (*entity.User, error) {
	var record entity.User
	if err := ctx.DB().Where("address=?", address).Take(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *userRepository) GetByServiceUserID(
	ctx xcontext.Context, service, serviceUserID string,
) (*entity.User, error) {
	var record entity.User
	err := ctx.DB().
		Model(&entity.User{}).
		Where("oauth2.service=? AND oauth2.service_user_id=?", service, serviceUserID).
		Joins("join oauth2 on users.id=oauth2.user_id").
		Take(&record).Error
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (r *userRepository) DeleteByID(ctx xcontext.Context, id string) error {
	panic("not implemented") // TODO: Implement
}

func (r *userRepository) UpsertByID(ctx xcontext.Context, id string, data *entity.User) error {
	var record entity.User
	err := ctx.DB().
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
