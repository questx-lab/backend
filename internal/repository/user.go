package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type UserRepository interface {
	Create(ctx xcontext.Context, data *entity.User) error
	UpdateByID(ctx xcontext.Context, id string, data *entity.User) error
	GetByID(ctx xcontext.Context, id string) (*entity.User, error)
	GetByName(ctx xcontext.Context, name string) (*entity.User, error)
	GetByIDs(ctx xcontext.Context, ids []string) ([]entity.User, error)
	GetByAddress(ctx xcontext.Context, address string) (*entity.User, error)
	GetByServiceUserID(ctx xcontext.Context, service, serviceUserID string) (*entity.User, error)
	GetByReferralCode(ctx xcontext.Context, referralCode string) (*entity.User, error)
	Count(ctx xcontext.Context) (int64, error)
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
	updateMap := map[string]any{}
	if data.Name != "" {
		updateMap["name"] = data.Name
		updateMap["is_new_user"] = false
	}

	if data.ProfilePictures != nil {
		updateMap["profile_pictures"] = data.ProfilePictures
	}

	if data.Address.Valid {
		updateMap["address"] = data.Address
	}

	return ctx.DB().Model(&entity.User{}).Where("id=?", id).Updates(updateMap).Error
}

func (r *userRepository) GetByID(ctx xcontext.Context, id string) (*entity.User, error) {
	var record entity.User
	if err := ctx.DB().Where("id=?", id).Take(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

func (r *userRepository) GetByName(ctx xcontext.Context, name string) (*entity.User, error) {
	var record entity.User
	if err := ctx.DB().Where("name=?", name).Take(&record).Error; err != nil {
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

func (r *userRepository) GetByReferralCode(ctx xcontext.Context, referralCode string) (*entity.User, error) {
	var record entity.User
	if err := ctx.DB().Where("referral_code=?", referralCode).Take(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *userRepository) Count(ctx xcontext.Context) (int64, error) {
	var count int64
	if err := ctx.DB().Model(&entity.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
