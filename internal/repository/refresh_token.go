package repository

import (
	"context"
	"errors"
	"time"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(context.Context, *entity.RefreshToken) error
	Rotate(ctx context.Context, family string) error
	Get(ctx context.Context, family string) (*entity.RefreshToken, error)
	Delete(ctx context.Context, family string) error
}

type refreshTokenRepository struct{}

func NewRefreshTokenRepository() *refreshTokenRepository {
	return &refreshTokenRepository{}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	return xcontext.DB(ctx).Create(token).Error
}

func (r *refreshTokenRepository) Rotate(ctx context.Context, family string) error {
	tx := xcontext.DB(ctx).Model(&entity.RefreshToken{}).
		Where("family=?", family).
		Updates(map[string]any{
			"counter":    gorm.Expr("counter+1"),
			"expiration": time.Now().Add(xcontext.Configs(ctx).Auth.RefreshToken.Expiration),
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

func (r *refreshTokenRepository) Get(ctx context.Context, family string) (*entity.RefreshToken, error) {
	var result entity.RefreshToken
	err := xcontext.DB(ctx).
		Joins("join users on users.id=refresh_tokens.user_id").
		Take(&result, "refresh_tokens.family=?", family).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *refreshTokenRepository) Delete(ctx context.Context, family string) error {
	return xcontext.DB(ctx).Delete(&entity.RefreshToken{}, "family=?", family).Error
}
