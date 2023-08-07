package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/domain/search"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type GetListCommunityFilter struct {
	Q                 string
	ReferredBy        string
	ReferralStatus    []entity.ReferralStatusType
	ByTrending        bool
	Status            entity.CommunityStatus
	OrderByReferredBy bool
}

type CommunityRepository interface {
	Create(ctx context.Context, e *entity.Community) error
	GetList(ctx context.Context, filter GetListCommunityFilter) ([]entity.Community, error)
	GetByID(ctx context.Context, id string) (*entity.Community, error)
	GetByHandle(ctx context.Context, handle string) (*entity.Community, error)
	UpdateByID(ctx context.Context, id string, e entity.Community) error
	GetByIDs(ctx context.Context, ids []string) ([]entity.Community, error)
	GetByHandles(ctx context.Context, handles []string) ([]entity.Community, error)
	UpdateReferralStatusByID(ctx context.Context, id string, status entity.ReferralStatusType) error
	DeleteByID(ctx context.Context, id string) error
	GetFollowingList(ctx context.Context, userID string, offset, limit int) ([]entity.Community, error)
	IncreaseFollowers(ctx context.Context, communityID string) error
	UpdateTrendingScore(ctx context.Context, communityID string, score int) error
}

type communityRepository struct {
	searchCaller client.SearchCaller
	redisClient  xredis.Client
}

func NewCommunityRepository(searchClient client.SearchCaller, redisClient xredis.Client) CommunityRepository {
	return &communityRepository{searchCaller: searchClient, redisClient: redisClient}
}

func (r *communityRepository) Create(ctx context.Context, e *entity.Community) error {
	if err := xcontext.DB(ctx).Model(e).Create(e).Error; err != nil {
		return err
	}

	if e.Status == entity.CommunityActive {
		err := r.searchCaller.IndexCommunity(ctx, e.ID, search.CommunityData{
			Handle:       e.Handle,
			DisplayName:  e.DisplayName,
			Introduction: string(e.Introduction),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *communityRepository) GetList(ctx context.Context, filter GetListCommunityFilter) ([]entity.Community, error) {
	if filter.Q == "" {
		var result []entity.Community
		tx := xcontext.DB(ctx)

		if filter.ByTrending {
			tx = tx.Order("trending_score DESC")
		}

		if filter.OrderByReferredBy {
			tx = tx.Where("referred_by IS NOT NULL").Order("referred_by, created_at ASC")
		}

		if filter.ReferredBy != "" {
			tx = tx.Where("referred_by=?", filter.ReferredBy)
		}

		if len(filter.ReferralStatus) != 0 {
			tx = tx.Where("referral_status IN (?)", filter.ReferralStatus)
		}

		if filter.Status != "" {
			tx = tx.Where("status=?", filter.Status)
		}

		if err := tx.Find(&result).Error; err != nil {
			return nil, err
		}

		return result, nil
	} else {
		ids, err := r.searchCaller.SearchCommunity(ctx, filter.Q)
		if err != nil {
			return nil, err
		}

		communities, err := r.GetByIDs(ctx, ids)
		if err != nil {
			return nil, err
		}

		communitySet := map[string]entity.Community{}
		for _, c := range communities {
			communitySet[c.ID] = c
		}

		orderedCommunities := []entity.Community{}
		for _, id := range ids {
			orderedCommunities = append(orderedCommunities, communitySet[id])
		}

		return orderedCommunities, nil
	}
}

func (r *communityRepository) GetByID(ctx context.Context, id string) (*entity.Community, error) {
	var record entity.Community
	err := r.redisClient.GetObj(ctx, r.cacheKey(id), &record)
	if err != nil && err != redis.Nil {
		xcontext.Logger(ctx).Errorf("Cannot get community from redis: %v", err)
	}

	if err == nil {
		return &record, nil
	}

	if err := xcontext.DB(ctx).Take(&record, "id=?", id).Error; err != nil {
		return nil, err
	}

	err = r.redisClient.SetObj(ctx, r.cacheKey(id), record, xcontext.Configs(ctx).Cache.TTL)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot set cache for community: %v", err)
	}

	return &record, nil
}

func (r *communityRepository) GetByHandle(ctx context.Context, handle string) (*entity.Community, error) {
	result := &entity.Community{}
	if err := xcontext.DB(ctx).Take(result, "handle=?", handle).Error; err != nil {
		return nil, err
	}

	err := r.redisClient.SetObj(ctx, r.cacheKey(result.ID), result, xcontext.Configs(ctx).Cache.TTL)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot set cache for community: %v", err)
	}

	return result, nil
}

func (r *communityRepository) GetByIDs(ctx context.Context, ids []string) ([]entity.Community, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	keys := []string{}
	for _, id := range ids {
		keys = append(keys, r.cacheKey(id))
	}

	var records []entity.Community
	notCacheIDs := []string{}

	values, err := r.redisClient.MGet(ctx, keys...)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot multiple get community from redis: %v", err)
		notCacheIDs = append(notCacheIDs, keys...)
	} else {
		for i := range keys {
			if values[i] != nil {
				s, ok := values[i].(string)
				if !ok {
					xcontext.Logger(ctx).Warnf("Invalid type of community %T", values[i])
				} else {
					var result entity.Community
					if err := json.Unmarshal([]byte(s), &result); err != nil {
						xcontext.Logger(ctx).Warnf("Cannot unmarshal community object: %v", err)
					} else {
						records = append(records, result)
						continue
					}
				}
			}

			notCacheIDs = append(notCacheIDs, ids[i])
		}
	}

	if len(notCacheIDs) != 0 {
		var dbRecords []entity.Community
		if err := xcontext.DB(ctx).Find(&dbRecords, "id IN (?)", ids).Error; err != nil {
			return nil, err
		}

		records = append(records, dbRecords...)
		redisKV := map[string]any{}
		for _, record := range dbRecords {
			redisKV[r.cacheKey(record.ID)] = record
		}

		if err := r.redisClient.MSet(ctx, redisKV); err != nil {
			xcontext.Logger(ctx).Warnf("Cannot multiple set for community redis: %v", err)
		}
	}

	return records, nil
}

func (r *communityRepository) GetByHandles(ctx context.Context, handles []string) ([]entity.Community, error) {
	result := []entity.Community{}
	tx := xcontext.DB(ctx)

	if tx.Find(&result, "handle IN (?)", handles).Error != nil {
		return nil, tx.Error
	}

	if len(result) != len(handles) {
		return nil, fmt.Errorf("got %d records, but expected %d", len(result), len(handles))
	}

	redisKV := map[string]any{}
	for _, record := range result {
		redisKV[r.cacheKey(record.ID)] = record
	}

	if err := r.redisClient.MSet(ctx, redisKV); err != nil {
		xcontext.Logger(ctx).Warnf("Cannot multiple set for community redis: %v", err)
	}

	return result, nil
}

func (r *communityRepository) UpdateByID(ctx context.Context, id string, e entity.Community) error {
	if err := r.redisClient.Del(ctx, r.cacheKey(id)); err != nil {
		return err
	}

	tx := xcontext.DB(ctx).
		Model(&entity.Community{}).
		Where("id=?", id).
		Omit("created_by", "created_at", "id").
		Updates(e)
	if err := tx.Error; err != nil {
		return err
	}

	if e.Introduction != nil || e.Handle != "" || e.Status == entity.CommunityActive {
		community, err := r.GetByID(ctx, id)
		if err != nil {
			return err
		}

		err = r.searchCaller.IndexCommunity(ctx, id, search.CommunityData{
			Handle:       community.Handle,
			DisplayName:  community.DisplayName,
			Introduction: string(community.Introduction),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *communityRepository) UpdateReferralStatusByID(
	ctx context.Context, id string, status entity.ReferralStatusType,
) error {
	if err := r.redisClient.Del(ctx, r.cacheKey(id)); err != nil {
		return err
	}

	tx := xcontext.DB(ctx).
		Model(&entity.Community{}).
		Where("id=?", id).
		Update("referral_status", status)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return errors.New("row affected is empty")
	}

	return nil
}

func (r *communityRepository) DeleteByID(ctx context.Context, id string) error {
	if err := r.redisClient.Del(ctx, r.cacheKey(id)); err != nil {
		return err
	}

	tx := xcontext.DB(ctx).
		Delete(&entity.Community{}, "id=?", id)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row affected is empty")
	}

	err := r.searchCaller.DeleteCommunity(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *communityRepository) GetFollowingList(ctx context.Context, userID string, offset, limit int) ([]entity.Community, error) {
	var result []entity.Community
	if err := xcontext.DB(ctx).
		Joins("join followers on communities.id = followers.community_id").
		Where("followers.user_id=?", userID).
		Limit(limit).Offset(offset).Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *communityRepository) IncreaseFollowers(ctx context.Context, communityID string) error {
	if err := r.redisClient.Del(ctx, r.cacheKey(communityID)); err != nil {
		return err
	}

	tx := xcontext.DB(ctx).
		Model(&entity.Community{}).
		Where("id=?", communityID).
		Update("followers", gorm.Expr("followers+1"))

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected > 1 {
		return errors.New("the number of affected rows is invalid")
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *communityRepository) UpdateTrendingScore(ctx context.Context, communityID string, score int) error {
	if err := r.redisClient.Del(ctx, r.cacheKey(communityID)); err != nil {
		return err
	}

	return xcontext.DB(ctx).
		Model(&entity.Community{}).
		Where("id=?", communityID).
		Update("trending_score", score).Error
}

func (r *communityRepository) cacheKey(communityID string) string {
	return fmt.Sprintf("cache:community:%s", communityID)
}
