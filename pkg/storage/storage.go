package storage

import "context"

type Storage interface {
	Upload(context.Context, *UploadObject) (*UploadResponse, error)
	BulkUpload(context.Context, []*UploadObject) ([]*UploadResponse, error)
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
