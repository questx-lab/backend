package storage

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_R2(t *testing.T) {
	storage := NewS3Storage(&S3Configs{
		Endpoint:  "https://589c81a53539a7457a5a0f50c8d9561c.r2.cloudflarestorage.com",
		AccessKey: "Cdf98e2603858ef94427be8a6f1d95e3",
		SecretKey: "a29df944635b922168883c1f1f47a72eefd7bb901ca375802649c35f344f8449",
		Region:    "us-west-1",
	})
	b, err := os.ReadFile("./test.jpg")
	require.NoError(t, err)
	ctx := context.Background()
	_, err = storage.Upload(ctx, &UploadObject{
		Bucket:   "testing",
		Prefix:   "image",
		FileName: "test",
		Mime:     "application/jpeg",
		Data:     b,
	})
	require.NoError(t, err)
}
