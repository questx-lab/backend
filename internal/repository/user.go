package repository

import (
	"context"
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
	"github.com/redis/go-redis/v9"
)

type UserRepository interface {
	Create(ctx context.Context, data *entity.User) error
	UpdateByID(ctx context.Context, id string, data *entity.User) error
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByName(ctx context.Context, name string) (*entity.User, error)
	GetByIDs(ctx context.Context, ids []string) ([]entity.User, error)
	GetByWalletAddress(ctx context.Context, walletAddress string) (*entity.User, error)
	GetByServiceUserID(ctx context.Context, service, serviceUserID string) (*entity.User, error)
	GetByReferralCode(ctx context.Context, referralCode string) (*entity.User, error)
	Count(ctx context.Context) (int64, error)
}

type userRepository struct {
	redisClient xredis.Client
}

func NewUserRepository(redisClient xredis.Client) UserRepository {
	return &userRepository{redisClient: redisClient}
}

func (r *userRepository) cacheKey(id string) string {
	return fmt.Sprintf("cache:user:%s", id)
}

func (r *userRepository) Create(ctx context.Context, data *entity.User) error {
	return xcontext.DB(ctx).Create(data).Error
}

func (r *userRepository) UpdateByID(ctx context.Context, id string, data *entity.User) error {
	if err := r.redisClient.Del(ctx, r.cacheKey(id)); err != nil {
		return err
	}

	updateMap := map[string]any{}
	if data.Name != "" {
		updateMap["name"] = data.Name
		updateMap["is_new_user"] = false
	}

	if data.ProfilePicture != "" {
		updateMap["profile_picture"] = data.ProfilePicture
	}

	if data.WalletAddress.Valid {
		updateMap["wallet_address"] = data.WalletAddress
	}

	return xcontext.DB(ctx).Model(&entity.User{}).Where("id=?", id).Updates(updateMap).Error
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	var record entity.User
	err := r.redisClient.GetObj(ctx, r.cacheKey(id), &record)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if err == nil {
		return &record, nil
	}

	if err := xcontext.DB(ctx).Where("id=?", id).Take(&record).Error; err != nil {
		return nil, err
	}

	err = r.redisClient.SetObj(ctx, r.cacheKey(id), record, xcontext.Configs(ctx).Cache.TTL)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot set cache for user: %v", err)
	}

	return &record, nil
}

func (r *userRepository) GetByName(ctx context.Context, name string) (*entity.User, error) {
	var record entity.User
	if err := xcontext.DB(ctx).Where("name=?", name).Take(&record).Error; err != nil {
		return nil, err
	}

	err := r.redisClient.SetObj(ctx, r.cacheKey(record.ID), record, xcontext.Configs(ctx).Cache.TTL)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot set cache for user: %v", err)
	}

	return &record, nil
}

func (r *userRepository) GetByIDs(ctx context.Context, ids []string) ([]entity.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var records []entity.User
	notCacheIDs := []string{}
	for _, id := range ids {
		var user entity.User
		err := r.redisClient.GetObj(ctx, r.cacheKey(id), &user)
		if err != nil && err != redis.Nil {
			return nil, err
		}

		if err == nil {
			records = append(records, user)
		} else {
			notCacheIDs = append(notCacheIDs, id)
		}
	}

	if len(notCacheIDs) != 0 {
		var dbRecords []entity.User
		if err := xcontext.DB(ctx).Where("id IN (?)", notCacheIDs).Find(&dbRecords).Error; err != nil {
			return nil, err
		}

		records = append(records, dbRecords...)
		for _, record := range dbRecords {
			err := r.redisClient.SetObj(ctx, r.cacheKey(record.ID), record, xcontext.Configs(ctx).Cache.TTL)
			if err != nil {
				xcontext.Logger(ctx).Warnf("Cannot set cache for user: %v", err)
			}
		}
	}

	return records, nil
}

func (r *userRepository) GetByWalletAddress(ctx context.Context, walletAddress string) (*entity.User, error) {
	var record entity.User
	if err := xcontext.DB(ctx).Where("wallet_address=?", walletAddress).Take(&record).Error; err != nil {
		return nil, err
	}

	err := r.redisClient.SetObj(ctx, r.cacheKey(record.ID), record, xcontext.Configs(ctx).Cache.TTL)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot set cache for user: %v", err)
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

	err = r.redisClient.SetObj(ctx, r.cacheKey(record.ID), record, xcontext.Configs(ctx).Cache.TTL)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot set cache for user: %v", err)
	}

	return &record, nil
}

func (r *userRepository) GetByReferralCode(ctx context.Context, referralCode string) (*entity.User, error) {
	var record entity.User
	if err := xcontext.DB(ctx).Where("referral_code=?", referralCode).Take(&record).Error; err != nil {
		return nil, err
	}

	err := r.redisClient.SetObj(ctx, r.cacheKey(record.ID), record, xcontext.Configs(ctx).Cache.TTL)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot set cache for user: %v", err)
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
