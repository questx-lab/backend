package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/utils"
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
	fields, values := utils.FieldMap(e)

	tableName := e.Table()
	fieldsStr := strings.Join(fields, ", ")
	placeHolder := utils.GeneratePlaceHolder(len(fields))

	stmt := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES(%s)`,
		tableName,
		fieldsStr,
		placeHolder,
	)
	if _, err := r.db.ExecContext(ctx, stmt, values...); err != nil {
		return fmt.Errorf("error insert project: %w", err)
	}
	return nil
}
