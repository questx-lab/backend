package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/router"
)

type OAuth2Repository interface {
	Create(ctx router.Context, data *entity.OAuth2) error
}

type oauth2Repository struct{}

func NewOAuth2Repository() OAuth2Repository {
	return &oauth2Repository{}
}

func (r *oauth2Repository) Create(ctx router.Context, data *entity.OAuth2) error {
	return ctx.DB().Create(data).Error
}
