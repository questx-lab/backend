package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type APIKeyRepository interface {
	Create(context.Context, *entity.APIKey) error
	Update(ctx context.Context, communityID, newKey string) error
	GetOwnerByKey(context.Context, string) (string, error)
	DeleteByCommunityID(context.Context, string) error
}

type apiKeyRepository struct{}

func NewAPIKeyRepository() *apiKeyRepository {
	return &apiKeyRepository{}
}

func (r *apiKeyRepository) Create(ctx context.Context, data *entity.APIKey) error {
	return xcontext.DB(ctx).Create(data).Error
}

func (r *apiKeyRepository) GetOwnerByKey(ctx context.Context, key string) (string, error) {
	var result entity.Community
	err := xcontext.DB(ctx).Model(&entity.Community{}).
		Select("communities.created_by").
		Joins("join api_keys on communities.id = api_keys.community_id").
		Take(&result, "api_keys.key=?", key).Error

	if err != nil {
		return "", err
	}

	return result.CreatedBy, nil
}

func (r *apiKeyRepository) DeleteByCommunityID(ctx context.Context, communityID string) error {
	return xcontext.DB(ctx).Delete(&entity.APIKey{}, "community_id=?", communityID).Error
}

func (r *apiKeyRepository) Update(ctx context.Context, communityID, newKey string) error {
	return xcontext.DB(ctx).Model(&entity.APIKey{}).
		Where("community_id=?", communityID).
		Update("key", newKey).Error
}
