package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/questx-lab/backend/internal/entity"
)

type ProjectRepository interface {
	Create(context.Context, *entity.Project) error
}
type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, e *entity.Project) error {
	tx := r.db.Table(e.Table()).Create(e)
	if err := tx.Error; err != nil {
		return err
	}
	return nil
}
