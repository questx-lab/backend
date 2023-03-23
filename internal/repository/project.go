package repository

import (
	"context"
	"time"

	"github.com/questx-lab/backend/internal/entity"

	"gorm.io/gorm"
)

type ProjectRepository interface {
	Create(ctx context.Context, e *entity.Project) error
	GetList(ctx context.Context, offset, limit int) ([]*entity.Project, error)
	GeyByID(ctx context.Context, id string) (*entity.Project, error)
	UpdateByID(ctx context.Context, id string, e *entity.Project) error
	DeleteByID(ctx context.Context, id string) error
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{
		db: db,
	}
}

func (r *projectRepository) Create(ctx context.Context, e *entity.Project) error {
	if err := r.db.Model(e).Create(e).Error; err != nil {
		return err
	}

	return nil
}

func (r *projectRepository) GetList(ctx context.Context, offset int, limit int) ([]*entity.Project, error) {
	var result []*entity.Project
	if err := r.db.Model(&entity.Project{}).Limit(limit).Offset(offset).Find(result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) GeyByID(ctx context.Context, id string) (*entity.Project, error) {
	result := &entity.Project{}
	if err := r.db.Model(&entity.Project{}).First(result, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return result, nil
}

func (r *projectRepository) UpdateByID(ctx context.Context, id string, e *entity.Project) error {
	if err := r.db.
		Model(&entity.Project{}).
		Where("id = ?", id).
		Omit("created_by", "created_at", "id").
		Updates(*e).Error; err != nil {
		return err
	}

	return nil
}

func (r *projectRepository) DeleteByID(ctx context.Context, id string) error {
	if err := r.db.
		Model(&entity.Project{}).
		Where("id = ?", id).
		Update("updated_at", time.Now()).Error; err != nil {
		return err
	}

	return nil
}
