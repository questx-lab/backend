package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/questx-lab/backend/internal/domain/search"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type GetListCommunityFilter struct {
	Q              string
	ReferredBy     string
	ReferralStatus entity.ReferralStatusType
	Offset         int
	Limit          int
	ByTrending     bool
}

type CommunityRepository interface {
	Create(ctx context.Context, e *entity.Community) error
	GetList(ctx context.Context, filter GetListCommunityFilter) ([]entity.Community, error)
	GetByID(ctx context.Context, id string) (*entity.Community, error)
	GetByName(ctx context.Context, name string) (*entity.Community, error)
	UpdateByID(ctx context.Context, id string, e entity.Community) error
	GetByIDs(ctx context.Context, ids []string) ([]entity.Community, error)
	UpdateReferralStatusByIDs(ctx context.Context, ids []string, status entity.ReferralStatusType) error
	DeleteByID(ctx context.Context, id string) error
	GetFollowingList(ctx context.Context, userID string, offset, limit int) ([]entity.Community, error)
	IncreaseFollowers(ctx context.Context, communityID string) error
	UpdateTrendingScore(ctx context.Context, communityID string, score int) error
}

type communityRepository struct {
	searchCaller search.Caller
}

func NewCommunityRepository(searchClient search.Caller) CommunityRepository {
	return &communityRepository{searchCaller: searchClient}
}

func (r *communityRepository) Create(ctx context.Context, e *entity.Community) error {
	if err := xcontext.DB(ctx).Model(e).Create(e).Error; err != nil {
		return err
	}

	err := r.searchCaller.IndexCommunity(ctx, e.ID, search.CommunityData{
		Name:         e.Name,
		Introduction: string(e.Introduction),
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *communityRepository) GetList(ctx context.Context, filter GetListCommunityFilter) ([]entity.Community, error) {
	if filter.Q == "" {
		var result []entity.Community
		tx := xcontext.DB(ctx).
			Limit(filter.Limit).
			Offset(filter.Offset)

		if filter.ByTrending {
			tx = tx.Order("trending_score DESC")
		}

		if filter.ReferredBy != "" {
			tx = tx.Where("referred_by=?", filter.ReferredBy)
		}

		if filter.ReferralStatus != "" {
			tx = tx.Where("referral_status=?", filter.ReferralStatus)
		}

		if err := tx.Find(&result).Error; err != nil {
			return nil, err
		}

		return result, nil
	} else {
		ids, err := r.searchCaller.SearchCommunity(ctx, filter.Q, filter.Offset, filter.Limit)
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
	result := &entity.Community{}
	if err := xcontext.DB(ctx).Take(result, "id=?", id).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *communityRepository) GetByName(ctx context.Context, name string) (*entity.Community, error) {
	result := &entity.Community{}
	if err := xcontext.DB(ctx).Take(result, "name=?", name).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *communityRepository) GetByIDs(ctx context.Context, ids []string) ([]entity.Community, error) {
	result := []entity.Community{}
	tx := xcontext.DB(ctx)

	if tx.Find(&result, "id IN (?)", ids).Error != nil {
		return nil, tx.Error
	}

	if len(result) != len(ids) {
		return nil, fmt.Errorf("got %d records, but expected %d", len(result), len(ids))
	}

	return result, nil
}

func (r *communityRepository) UpdateByID(ctx context.Context, id string, e entity.Community) error {
	tx := xcontext.DB(ctx).
		Model(&entity.Community{}).
		Where("id=?", id).
		Omit("created_by", "created_at", "id").
		Updates(e)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return fmt.Errorf("row affected is empty")
	}

	err := r.searchCaller.ReplaceCommunity(ctx, e.ID, search.CommunityData{
		Name:         e.Name,
		Introduction: string(e.Introduction),
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *communityRepository) UpdateReferralStatusByIDs(
	ctx context.Context, ids []string, status entity.ReferralStatusType,
) error {
	tx := xcontext.DB(ctx).
		Model(&entity.Community{}).
		Where("id IN (?)", ids).
		Update("referral_status", status)
	if err := tx.Error; err != nil {
		return err
	}

	if tx.RowsAffected == 0 {
		return errors.New("row affected is empty")
	}

	if int(tx.RowsAffected) != len(ids) {
		return fmt.Errorf("got %d row affected, but expected %d", tx.RowsAffected, len(ids))
	}

	return nil
}

func (r *communityRepository) DeleteByID(ctx context.Context, id string) error {
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
	return xcontext.DB(ctx).
		Model(&entity.Community{}).
		Where("id=?", communityID).
		Update("trending_score", score).Error
}
