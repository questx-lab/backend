package repository

import (
	"context"
	"database/sql"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/utils/database"
)

type ProjectRepository interface {
	Create(context.Context, *entity.Project) error
}
type projectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(ctx context.Context, e *entity.Project) error {
	return database.Insert(ctx, r.db, e)
}
