package storage

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3Storage struct {
	uploader *s3manager.Uploader
	cfg      *S3Configs
}

type S3Configs struct {
	Region   string
	Endpoint string

	AccessKey string
	SecretKey string
}

func NewS3Storage(cfg *S3Configs) Storage {
	session, _ := session.NewSession(&aws.Config{
		Region:           aws.String(cfg.Region),
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
		Endpoint:         aws.String(cfg.Endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})

	return &s3Storage{
		uploader: s3manager.NewUploader(session),
		cfg:      cfg,
	}

}

func (s *s3Storage) generateUploadURL(object *UploadObject) *UploadResponse {
	h := md5.New()
	io.WriteString(h, string(object.Data))
	fileName := fmt.Sprintf("%s/%x.png", object.Prefix, h.Sum(nil))

	return &UploadResponse{
		Url:      fmt.Sprintf("%s/%s/%s", s.cfg.Endpoint, object.Bucket, fileName),
		FileName: fileName,
	}
}

func (s *s3Storage) Upload(ctx context.Context, object *UploadObject) (*UploadResponse, error) {
	resp := s.generateUploadURL(object)
	// Upload the file to S3.
	_, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:      aws.String(object.Bucket),
		Key:         aws.String(resp.FileName),
		Body:        bytes.NewReader(object.Data),
		ACL:         aws.String("public-read"),
		ContentType: aws.String(object.Mime),
	})
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w, bucket %s, key %s", err, object.Bucket, resp.Url)
	}
	return resp, nil
}

func (s *s3Storage) BulkUpload(_ context.Context, _ []*UploadObject) ([]*UploadResponse, error) {
	panic("not implemented") // TODO: Implement
}
