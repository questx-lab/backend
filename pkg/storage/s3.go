package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"github.com/questx-lab/backend/config"
)

type s3Storage struct {
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	cfg        config.S3Configs
}

func NewS3Storage(cfg config.S3Configs) Storage {
	session, _ := session.NewSession(&aws.Config{
		Region:           aws.String(cfg.Region),
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
		Endpoint:         aws.String(cfg.Endpoint),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(cfg.SSLDisabled),
	})

	return &s3Storage{
		uploader:   s3manager.NewUploader(session),
		downloader: s3manager.NewDownloader(session),
		cfg:        cfg,
	}

}

func (s *s3Storage) generateUploadURL(ctx context.Context, object *UploadObject) *UploadResponse {
	id := uuid.NewString()
	fileName := fmt.Sprintf("%s/%s-%s", object.Prefix, id, object.FileName)

	return &UploadResponse{
		Url:      fmt.Sprintf("%s/%s/%s", s.cfg.PublicEndpoint, object.Bucket, fileName),
		FileName: fileName,
	}
}

func (s *s3Storage) Download(ctx context.Context, bucket, item string) ([]byte, error) {
	buffer := aws.NewWriteAtBuffer(nil)
	_, err := s.downloader.Download(buffer, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	})
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (s *s3Storage) DownloadFromURL(ctx context.Context, rawURL string) ([]byte, error) {
	parts, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	bucket, item, found := strings.Cut(strings.TrimLeft(parts.Path, "/"), "/")
	if !found {
		return nil, errors.New("not found bucket and item in url")
	}

	return s.Download(ctx, strings.Trim(bucket, "/"), strings.Trim(item, "/"))
}

func (s *s3Storage) Upload(ctx context.Context, object *UploadObject) (*UploadResponse, error) {
	resp := s.generateUploadURL(ctx, object)
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

func (s *s3Storage) BulkUpload(ctx context.Context, objects []*UploadObject) ([]*UploadResponse, error) {
	bObjects := make([]s3manager.BatchUploadObject, 0, len(objects))
	out := make([]*UploadResponse, 0, len(objects))
	for _, o := range objects {
		resp := s.generateUploadURL(ctx, o)
		b := s3manager.BatchUploadObject{
			Object: &s3manager.UploadInput{
				Bucket:      aws.String(o.Bucket),
				Key:         aws.String(resp.FileName),
				Body:        bytes.NewReader(o.Data),
				ACL:         aws.String("public-read"),
				ContentType: aws.String(o.Mime),
			},
		}
		bObjects = append(bObjects, b)
		out = append(out, resp)
	}

	// Upload the file to S3.
	if err := s.uploader.UploadWithIterator(ctx, &s3manager.UploadObjectsIterator{
		Objects: bObjects,
	}); err != nil {
		return nil, err
	}
	return out, nil
}
