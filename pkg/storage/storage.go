package storage

import (
	"github.com/questx-lab/backend/pkg/xcontext"
)

type Storage interface {
	Upload(xcontext.Context, *UploadObject) (*UploadResponse, error)
	BulkUpload(xcontext.Context, []*UploadObject) ([]*UploadResponse, error)
}

type UploadObject struct {
	Bucket   string
	Prefix   string
	FileName string
	Mime     string
	Data     []byte
}

type UploadResponse struct {
	Url      string
	FileName string
}
