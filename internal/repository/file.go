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

func NewFileRepository() FileRepository {
	return &fileRepository{}
}

func (r *fileRepository) Create(ctx xcontext.Context, e *entity.File) error {
	if err := ctx.DB().Create(e).Error; err != nil {
		return err
	}
	return nil
}

func (r *fileRepository) BulkInsert(_ xcontext.Context, _ []*entity.File) error {
	panic("not implemented") // TODO: Implement
}

func (r *fileRepository) GetByID(_ xcontext.Context, _ string) (*entity.File, error) {
	panic("not implemented") // TODO: Implement
}
