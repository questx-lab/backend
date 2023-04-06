package storage

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
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
	Env       string
}

func NewS3Storage(cfg *S3Configs) Storage {
	disableSSL := true
	if cfg.Env != "local" {
		disableSSL = false
	}
	session, _ := session.NewSession(&aws.Config{
		Region:           aws.String(cfg.Region),
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
		Endpoint:         aws.String(cfg.Endpoint),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(disableSSL),
	})

	return &s3Storage{
		uploader: s3manager.NewUploader(session),
		cfg:      cfg,
	}

}

func (s *s3Storage) generateUploadURL(object *UploadObject) *UploadResponse {
	id := uuid.NewString()
	fileName := fmt.Sprintf("%s/%s-%s", object.Prefix, id, object.FileName)

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
		log.Println(err)
		return nil, fmt.Errorf("upload failed: %w, bucket %s, key %s", err, object.Bucket, resp.Url)
	}
	return resp, nil
}

func (s *s3Storage) BulkUpload(_ context.Context, _ []*UploadObject) ([]*UploadResponse, error) {
	panic("not implemented") // TODO: Implement
}
