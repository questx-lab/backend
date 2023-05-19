package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type APIKeyRepository interface {
	Create(context.Context, *entity.APIKey) error
	Update(ctx context.Context, projectID, newKey string) error
	GetOwnerByKey(context.Context, string) (string, error)
	DeleteByProjectID(context.Context, string) error
}

type apiKeyRepository struct{}

func NewAPIKeyRepository() *apiKeyRepository {
	return &apiKeyRepository{}
}

func (r *apiKeyRepository) Create(ctx context.Context, data *entity.APIKey) error {
	return xcontext.DB(ctx).Create(data).Error
}

func (r *apiKeyRepository) GetOwnerByKey(ctx context.Context, key string) (string, error) {
	var result entity.Project
	err := xcontext.DB(ctx).Model(&entity.Project{}).
		Select("projects.created_by").
		Joins("join api_keys on projects.id = api_keys.project_id").
		Take(&result, "api_keys.key=?", key).Error

	if err != nil {
		return "", err
	}

	return result.CreatedBy, nil
}

func (r *apiKeyRepository) DeleteByProjectID(ctx context.Context, projectID string) error {
	return xcontext.DB(ctx).Delete(&entity.APIKey{}, "project_id=?", projectID).Error
}

func (r *apiKeyRepository) Update(ctx context.Context, projectID, newKey string) error {
	return xcontext.DB(ctx).Model(&entity.APIKey{}).
		Where("project_id=?", projectID).
		Update("key", newKey).Error
}
