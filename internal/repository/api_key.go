package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type APIKeyRepository interface {
	Create(xcontext.Context, *entity.APIKey) error
	Update(ctx xcontext.Context, projectID, newKey string) error
	GetOwnerByKey(xcontext.Context, string) (string, error)
	DeleteByProjectID(xcontext.Context, string) error
}

type apiKeyRepository struct{}

func NewAPIKeyRepository() *apiKeyRepository {
	return &apiKeyRepository{}
}

func (r *apiKeyRepository) Create(ctx xcontext.Context, data *entity.APIKey) error {
	return ctx.DB().Create(data).Error
}

func (r *apiKeyRepository) GetOwnerByKey(ctx xcontext.Context, key string) (string, error) {
	var result entity.Project
	err := ctx.DB().Model(&entity.Project{}).
		Select("projects.created_by").
		Joins("join api_keys on projects.id = api_keys.project_id").
		Take(&result, "api_keys.key=?", key).Error

	if err != nil {
		return "", err
	}

	return result.CreatedBy, nil
}

func (r *apiKeyRepository) DeleteByProjectID(ctx xcontext.Context, projectID string) error {
	return ctx.DB().Delete(&entity.APIKey{}, "project_id=?", projectID).Error
}

func (r *apiKeyRepository) Update(ctx xcontext.Context, projectID, newKey string) error {
	return ctx.DB().Model(&entity.APIKey{}).
		Where("project_id=?", projectID).
		Update("key", newKey).Error
}
