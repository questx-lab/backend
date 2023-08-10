package repository

import (
	"context"
	"encoding/json"
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

func (r *userRepository) cache(ctx context.Context, users ...entity.User) {
	redisKV := map[string]any{}
	for _, record := range users {
		redisKV[r.cacheKey(record.ID)] = record
	}

	if err := r.redisClient.MSet(ctx, redisKV); err != nil {
		xcontext.Logger(ctx).Warnf("Cannot multiple set for user redis: %v", err)
	}
}

func (r *userRepository) fromCache(ctx context.Context, ids ...string) []entity.User {
	keys := []string{}
	for _, id := range ids {
		keys = append(keys, r.cacheKey(id))
	}

	var records []entity.User
	values, err := r.redisClient.MGet(ctx, keys...)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot multiple get user from redis: %v", err)
		return nil
	}

	for i := range keys {
		if values[i] == nil {
			continue
		}

		s, ok := values[i].(string)
		if !ok {
			xcontext.Logger(ctx).Warnf("Invalid type of user %T", values[i])
			continue
		}

		var result entity.User
		if err := json.Unmarshal([]byte(s), &result); err != nil {
			xcontext.Logger(ctx).Warnf("Cannot unmarshal user object: %v", err)
			continue
		}

		records = append(records, result)
	}

	return records
}

func (r *userRepository) invalidateCache(ctx context.Context, ids ...string) {
	keys := []string{}
	for _, id := range ids {
		keys = append(keys, r.cacheKey(id))
	}

	if err := r.redisClient.Del(ctx, keys...); err != nil && err != redis.Nil {
		xcontext.Logger(ctx).Warnf("Cannot invalidate redis key: %v", err)
	}
}

func (r *userRepository) Create(ctx context.Context, data *entity.User) error {
	return xcontext.DB(ctx).Create(data).Error
}

func (r *userRepository) UpdateByID(ctx context.Context, id string, data *entity.User) error {
	r.invalidateCache(ctx, id)

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
	if users := r.fromCache(ctx, id); len(users) > 0 {
		return &users[0], nil
	}

	var record entity.User
	if err := xcontext.DB(ctx).Where("id=?", id).Take(&record).Error; err != nil {
		return nil, err
	}

	r.cache(ctx, record)
	return &record, nil
}

func (r *userRepository) GetByName(ctx context.Context, name string) (*entity.User, error) {
	var record entity.User
	if err := xcontext.DB(ctx).Where("name=?", name).Take(&record).Error; err != nil {
		return nil, err
	}

	r.cache(ctx, record)
	return &record, nil
}

func (r *userRepository) GetByIDs(ctx context.Context, ids []string) ([]entity.User, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	records := r.fromCache(ctx, ids...)
	notCacheIDs := []string{}
	for _, id := range ids {
		isCached := false
		for _, cachedUser := range records {
			if id == cachedUser.ID {
				isCached = true
				break
			}
		}

		if !isCached {
			notCacheIDs = append(notCacheIDs, id)
		}
	}

	if len(notCacheIDs) != 0 {
		var dbRecords []entity.User
		if err := xcontext.DB(ctx).Where("id IN (?)", notCacheIDs).Find(&dbRecords).Error; err != nil {
			return nil, err
		}

		r.cache(ctx, dbRecords...)
	}

	return records, nil
}

func (r *userRepository) GetByWalletAddress(ctx context.Context, walletAddress string) (*entity.User, error) {
	var record entity.User
	if err := xcontext.DB(ctx).Where("wallet_address=?", walletAddress).Take(&record).Error; err != nil {
		return nil, err
	}

	r.cache(ctx, record)
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

	r.cache(ctx, record)
	return &record, nil
}

func (r *userRepository) GetByReferralCode(ctx context.Context, referralCode string) (*entity.User, error) {
	var record entity.User
	if err := xcontext.DB(ctx).Where("referral_code=?", referralCode).Take(&record).Error; err != nil {
		return nil, err
	}

	r.cache(ctx, record)
	return &record, nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := xcontext.DB(ctx).Model(&entity.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
