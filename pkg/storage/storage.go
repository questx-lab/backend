package storage

import "context"

type Storage interface {
	Download(ctx context.Context, bucket, item string) ([]byte, error)
	DownloadFromURL(ctx context.Context, url string) ([]byte, error)
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
	Url      string `json:"url"`
	FileName string `json:"filename"`
}
