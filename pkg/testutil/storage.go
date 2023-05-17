package testutil

import (
	"context"

	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/storage"
)

type MockStorage struct {
	UploadFunc     func(context.Context, *storage.UploadObject) (*storage.UploadResponse, error)
	BulkUploadFunc func(context.Context, []*storage.UploadObject) ([]*storage.UploadResponse, error)
}

func (m *MockStorage) Upload(
	ctx context.Context, obj *storage.UploadObject,
) (*storage.UploadResponse, error) {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, obj)
	}

	return nil, errorx.New(errorx.NotImplemented, "Not implemented")
}

func (m *MockStorage) BulkUpload(
	ctx context.Context, obj []*storage.UploadObject,
) ([]*storage.UploadResponse, error) {
	if m.BulkUploadFunc != nil {
		return m.BulkUploadFunc(ctx, obj)
	}

	return nil, errorx.New(errorx.NotImplemented, "Not implemented")
}
