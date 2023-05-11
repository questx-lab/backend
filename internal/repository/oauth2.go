package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type OAuth2Repository interface {
	Create(ctx xcontext.Context, data *entity.OAuth2) error
	GetByUserID(ctx xcontext.Context, service, userID string) (*entity.OAuth2, error)
	GetAllByUserID(ctx xcontext.Context, userID string) ([]entity.OAuth2, error)
}

type oauth2Repository struct{}

func NewOAuth2Repository() OAuth2Repository {
	return &oauth2Repository{}
}

func (r *oauth2Repository) Create(ctx xcontext.Context, data *entity.OAuth2) error {
	return ctx.DB().Create(data).Error
}

func (r *oauth2Repository) GetByUserID(ctx xcontext.Context, service, userID string) (*entity.OAuth2, error) {
	var result entity.OAuth2
	err := ctx.DB().Take(&result, "service=? AND user_id=?", service, userID).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *oauth2Repository) GetAllByUserID(ctx xcontext.Context, userID string) ([]entity.OAuth2, error) {
	var result []entity.OAuth2
	err := ctx.DB().Find(&result, "user_id=?", userID).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}
