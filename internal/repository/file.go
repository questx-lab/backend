package repository

import (
	"context"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type FileRepository interface {
	Create(context.Context, *entity.File) error
	BulkInsert(context.Context, []*entity.File) error
	GetByID(context.Context, string) (*entity.File, error)
}

type fileRepository struct {
}

func NewFileRepository() FileRepository {
	return &fileRepository{}
}

func (r *fileRepository) Create(ctx context.Context, e *entity.File) error {
	if err := xcontext.DB(ctx).Create(e).Error; err != nil {
		return err
	}
	return nil
}

func (r *fileRepository) BulkInsert(ctx context.Context, es []*entity.File) error {
	if err := xcontext.DB(ctx).Create(es).Error; err != nil {
		return err
	}
	return nil
}

func (r *fileRepository) GetByID(_ context.Context, _ string) (*entity.File, error) {
	panic("not implemented") // TODO: Implement
}
