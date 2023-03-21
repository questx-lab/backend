package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"gorm.io/gorm"
)

type OAuth2Repository interface {
	Create(ctx context.Context, data *entity.OAuth2) error
}

type oauth2Repository struct {
	db *gorm.DB
}

func NewOAuth2Repository(db *gorm.DB) OAuth2Repository {
	return &oauth2Repository{db: db}
}

func (r *oauth2Repository) Create(ctx context.Context, data *entity.OAuth2) error {
	return r.db.Create(data).Error
}
