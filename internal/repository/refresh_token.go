package repository

import (
	"errors"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(xcontext.Context, *entity.RefreshToken) error
	Rotate(ctx xcontext.Context, family string) error
	Get(ctx xcontext.Context, family string) (*entity.RefreshToken, error)
	Delete(ctx xcontext.Context, family string) error
}

type refreshTokenRepository struct{}

func NewRefreshTokenRepository() *refreshTokenRepository {
	return &refreshTokenRepository{}
}

func (r *refreshTokenRepository) Create(ctx xcontext.Context, token *entity.RefreshToken) error {
	return ctx.DB().Create(token).Error
}

func (r *refreshTokenRepository) Rotate(ctx xcontext.Context, family string) error {
	tx := ctx.DB().Model(&entity.RefreshToken{}).
		Where("family=?", family).
		Updates(map[string]any{
			"counter":    gorm.Expr("counter+1"),
			"expiration": time.Now().Add(ctx.Configs().Auth.RefreshToken.Expiration),
		})

	if tx.Error != nil {
		return tx.Error
	}

	if tx.RowsAffected > 1 {
		return errors.New("the number of rows effected is invalid")
	}

	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *refreshTokenRepository) Get(ctx xcontext.Context, family string) (*entity.RefreshToken, error) {
	var result entity.RefreshToken
	err := ctx.DB().
		Joins("join users on users.id=refresh_tokens.user_id").
		Take(&result, "refresh_tokens.family=?", family).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *refreshTokenRepository) Delete(ctx xcontext.Context, family string) error {
	return ctx.DB().Delete(&entity.RefreshToken{}, "family=?", family).Error
}
