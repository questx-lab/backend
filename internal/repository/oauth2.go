package repository

import (
	"fmt"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type OAuth2Repository interface {
	Create(ctx xcontext.Context, data *entity.OAuth2) error
}

type oauth2Repository struct{}

func NewOAuth2Repository() OAuth2Repository {
	return &oauth2Repository{}
}

func (r *oauth2Repository) Create(ctx xcontext.Context, data *entity.OAuth2) error {
	data.ServiceUserID = fmt.Sprintf("%s_%s", data.Service, data.ServiceUserID)
	return ctx.DB().Create(data).Error
}
