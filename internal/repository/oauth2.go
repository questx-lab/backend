package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type OAuth2Repository interface {
	Create(ctx context.Context, data *entity.OAuth2) error
	GetByUserID(ctx context.Context, service, userID string) (*entity.OAuth2, error)
	GetAllByUserIDs(ctx context.Context, userID ...string) ([]entity.OAuth2, error)
}

type oauth2Repository struct{}

func NewOAuth2Repository() OAuth2Repository {
	return &oauth2Repository{}
}

func (r *oauth2Repository) Create(ctx context.Context, data *entity.OAuth2) error {
	return xcontext.DB(ctx).Create(data).Error
}

func (r *oauth2Repository) GetByUserID(ctx context.Context, service, userID string) (*entity.OAuth2, error) {
	var result entity.OAuth2
	err := xcontext.DB(ctx).Take(&result, "service=? AND user_id=?", service, userID).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *oauth2Repository) GetAllByUserIDs(ctx context.Context, userIDs ...string) ([]entity.OAuth2, error) {
	var result []entity.OAuth2
	err := xcontext.DB(ctx).Find(&result, "user_id IN (?)", userIDs).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}
