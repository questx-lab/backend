package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/client"
	"github.com/questx-lab/backend/internal/domain/search"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"github.com/questx-lab/backend/pkg/xredis"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	DecreaseFollowers(ctx context.Context, communityID string) error
	UpdateTrendingScore(ctx context.Context, communityID string, score int) error
	SetStats(ctx context.Context, record *entity.CommunityStats) error
	GetStats(ctx context.Context, communityID string, begin, end time.Time) ([]entity.CommunityStats, error)
	GetLastStat(ctx context.Context, communityID string) (*entity.CommunityStats, error)
}

type communityRepository struct {
	searchCaller client.SearchCaller
	redisClient  xredis.Client
}

func NewCommunityRepository(searchClient client.SearchCaller, redisClient xredis.Client) CommunityRepository {
	return &communityRepository{searchCaller: searchClient, redisClient: redisClient}
}

func (r *communityRepository) cacheKeyByID(communityID string) string {
	return fmt.Sprintf("cache:community:%s", communityID)
}

func (r *communityRepository) cacheKeyByHandle(communityHandle string) string {
	return fmt.Sprintf("cache:community:handle:%s", communityHandle)
}

func (r *communityRepository) cache(ctx context.Context, communities ...entity.Community) {
	redisKV := map[string]any{}
	for _, record := range communities {
		redisKV[r.cacheKeyByID(record.ID)] = record
		redisKV[r.cacheKeyByHandle(record.Handle)] = record
	}

	if err := r.redisClient.MSet(ctx, redisKV); err != nil {
		xcontext.Logger(ctx).Warnf("Cannot multiple set for community redis: %v", err)
	}
}

func (r *communityRepository) fromCacheByID(ctx context.Context, ids ...string) []entity.Community {
	keys := []string{}
	for _, id := range ids {
		keys = append(keys, r.cacheKeyByID(id))
	}

	var records []entity.Community
	values, err := r.redisClient.MGet(ctx, keys...)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot multiple get community from redis: %v", err)
		return nil
	}

	for i := range keys {
		if values[i] == nil {
			continue
		}

		s, ok := values[i].(string)
		if !ok {
			xcontext.Logger(ctx).Warnf("Invalid type of community %T", values[i])
			continue
		}

		var result entity.Community
		if err := json.Unmarshal([]byte(s), &result); err != nil {
			xcontext.Logger(ctx).Warnf("Cannot unmarshal community object: %v", err)
			continue
		}

		records = append(records, result)
	}

	return records
}

func (r *communityRepository) fromCacheByHandle(ctx context.Context, handles ...string) []entity.Community {
	keys := []string{}
	for _, handle := range handles {
		keys = append(keys, r.cacheKeyByHandle(handle))
	}

	var records []entity.Community
	values, err := r.redisClient.MGet(ctx, keys...)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot multiple get community from redis: %v", err)
		return nil
	}

	for i := range keys {
		if values[i] == nil {
			continue
		}

		s, ok := values[i].(string)
		if !ok {
			xcontext.Logger(ctx).Warnf("Invalid type of community %T", values[i])
			continue
		}

		var result entity.Community
		if err := json.Unmarshal([]byte(s), &result); err != nil {
			xcontext.Logger(ctx).Warnf("Cannot unmarshal community object: %v", err)
			continue
		}

		records = append(records, result)
	}

	return records
}

func (r *communityRepository) invalidateCache(ctx context.Context, ids ...string) {
	records := r.fromCacheByID(ctx, ids...)

	keys := []string{}
	for _, record := range records {
		keys = append(keys, r.cacheKeyByID(record.ID))
		keys = append(keys, r.cacheKeyByHandle(record.Handle))
	}

	if len(keys) > 0 {
		if err := r.redisClient.Del(ctx, keys...); err != nil && err != redis.Nil {
			xcontext.Logger(ctx).Warnf("Cannot invalidate community redis key: %v", err)
		}
	}
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
	if c := r.fromCacheByID(ctx, id); len(c) > 0 {
		return &c[0], nil
	}

	var record entity.Community
	if err := xcontext.DB(ctx).Take(&record, "id=?", id).Error; err != nil {
		return nil, err
	}

	r.cache(ctx, record)
	return &record, nil
}

func (r *communityRepository) GetByHandle(ctx context.Context, handle string) (*entity.Community, error) {
	if c := r.fromCacheByHandle(ctx, handle); len(c) > 0 {
		return &c[0], nil
	}

	result := entity.Community{}
	if err := xcontext.DB(ctx).Take(&result, "handle=?", handle).Error; err != nil {
		return nil, err
	}

	r.cache(ctx, result)
	return &result, nil
}

func (r *communityRepository) GetByIDs(ctx context.Context, ids []string) ([]entity.Community, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	records := r.fromCacheByID(ctx, ids...)
	notCacheIDs := []string{}
	for _, id := range ids {
		isCached := false
		for _, cachedCommunity := range records {
			if id == cachedCommunity.ID {
				isCached = true
				break
			}
		}

		if !isCached {
			notCacheIDs = append(notCacheIDs, id)
		}
	}

	if len(notCacheIDs) != 0 {
		var dbRecords []entity.Community
		if err := xcontext.DB(ctx).Find(&dbRecords, "id IN (?)", ids).Error; err != nil {
			return nil, err
		}

		records = append(records, dbRecords...)
		r.cache(ctx, dbRecords...)
	}

	return records, nil
}

func (r *communityRepository) GetByHandles(ctx context.Context, handles []string) ([]entity.Community, error) {
	records := r.fromCacheByHandle(ctx, handles...)
	notCacheIDs := []string{}
	for _, handle := range handles {
		isCached := false
		for _, cachedCommunity := range records {
			if handle == cachedCommunity.Handle {
				isCached = true
				break
			}
		}

		if !isCached {
			notCacheIDs = append(notCacheIDs, handle)
		}
	}

	if len(notCacheIDs) != 0 {
		dbRecords := []entity.Community{}
		if err := xcontext.DB(ctx).Find(&dbRecords, "handle IN (?)", handles).Error; err != nil {
			return nil, err
		}

		records = append(records, dbRecords...)
		r.cache(ctx, dbRecords...)
	}

	return records, nil
}

func (r *communityRepository) UpdateByID(ctx context.Context, id string, e entity.Community) error {
	r.invalidateCache(ctx, id)

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
	r.invalidateCache(ctx, id)

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
	r.invalidateCache(ctx, id)

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
	r.invalidateCache(ctx, communityID)

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
	r.invalidateCache(ctx, communityID)

	return xcontext.DB(ctx).
		Model(&entity.Community{}).
		Where("id=?", communityID).
		Update("trending_score", score).Error
}

func (r *communityRepository) DecreaseFollowers(ctx context.Context, communityID string) error {
	tx := xcontext.DB(ctx).
		Model(&entity.Community{}).
		Where("id=?", communityID).
		Update("followers", gorm.Expr("followers-1"))

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

func (r *communityRepository) SetStats(ctx context.Context, record *entity.CommunityStats) error {
	return xcontext.DB(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "community_id"},
				{Name: "date"},
			},
			DoUpdates: clause.Assignments(map[string]any{
				"follower_count": record.FollowerCount,
			}),
		}).Create(record).Error
}

func (r *communityRepository) GetStats(
	ctx context.Context, communityID string, begin, end time.Time,
) ([]entity.CommunityStats, error) {
	var result []entity.CommunityStats
	tx := xcontext.DB(ctx).Where("date>=? AND date<=?", begin, end)

	if communityID != "" {
		tx.Where("community_id=?", communityID)
	} else {
		tx.Where("community_id IS NULL")
	}

	if err := tx.Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *communityRepository) GetLastStat(ctx context.Context, communityID string) (*entity.CommunityStats, error) {
	tx := xcontext.DB(ctx).Model(&entity.CommunityStats{}).Order("date DESC")
	if communityID != "" {
		tx.Where("community_id=?", communityID)
	} else {
		tx.Where("community_id IS NULL")
	}

	var result entity.CommunityStats
	if err := tx.Take(&result).Error; err != nil {
		return nil, err
	}

	return &result, nil
}
