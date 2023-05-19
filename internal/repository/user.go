package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type UserRepository interface {
	Create(ctx context.Context, data *entity.User) error
	UpdateByID(ctx context.Context, id string, data *entity.User) error
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByName(ctx context.Context, name string) (*entity.User, error)
	GetByIDs(ctx context.Context, ids []string) ([]entity.User, error)
	GetByAddress(ctx context.Context, address string) (*entity.User, error)
	GetByServiceUserID(ctx context.Context, service, serviceUserID string) (*entity.User, error)
	GetByReferralCode(ctx context.Context, referralCode string) (*entity.User, error)
	Count(ctx context.Context) (int64, error)
}

type userRepository struct {
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) Create(ctx context.Context, data *entity.User) error {
	return xcontext.DB(ctx).Create(data).Error
}

func (r *userRepository) UpdateByID(ctx context.Context, id string, data *entity.User) error {
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

	return xcontext.DB(ctx).Model(&entity.User{}).Where("id=?", id).Updates(updateMap).Error
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	var record entity.User
	if err := xcontext.DB(ctx).Where("id=?", id).Take(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

func (r *userRepository) GetByName(ctx context.Context, name string) (*entity.User, error) {
	var record entity.User
	if err := xcontext.DB(ctx).Where("name=?", name).Take(&record).Error; err != nil {
		return nil, err
	}

	return &record, nil
}

func (r *userRepository) GetByIDs(ctx context.Context, ids []string) ([]entity.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var record []entity.User
	if err := xcontext.DB(ctx).Where("id IN (?)", ids).Find(&record).Error; err != nil {
		return nil, err
	}

	return record, nil
}

func (r *userRepository) GetByAddress(ctx context.Context, address string) (*entity.User, error) {
	var record entity.User
	if err := xcontext.DB(ctx).Where("address=?", address).Take(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *userRepository) GetByServiceUserID(
	ctx context.Context, service, serviceUserID string,
) (*entity.User, error) {
	var record entity.User
	err := xcontext.DB(ctx).
		Model(&entity.User{}).
		Where("oauth2.service=? AND oauth2.service_user_id=?", service, serviceUserID).
		Joins("join oauth2 on users.id=oauth2.user_id").
		Take(&record).Error
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (r *userRepository) GetByReferralCode(ctx context.Context, referralCode string) (*entity.User, error) {
	var record entity.User
	if err := xcontext.DB(ctx).Where("referral_code=?", referralCode).Take(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := xcontext.DB(ctx).Model(&entity.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
