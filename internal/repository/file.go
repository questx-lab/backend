package repository

import (
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type FileRepository interface {
	Create(xcontext.Context, *entity.File) error
	BulkInsert(xcontext.Context, []*entity.File) error
	GetByID(xcontext.Context, string) (*entity.File, error)
}

type fileRepository struct {
}
