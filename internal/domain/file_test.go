package domain

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/mocks"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/testutil"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func generateRandomImage(path string) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))
	img.Set(2, 3, color.RGBA{255, 0, 0, 255})
	f, _ := os.Create(path)
	defer f.Close()
	_ = png.Encode(f, img)
}

func deleteImage(path string) {
	_ = os.Remove(path)
}

func Test_fileDomain_UploadAvatar(t *testing.T) {

	path := "out.png"
	generateRandomImage(path)
	defer deleteImage(path)
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()
	fw, err := writer.CreateFormFile("avatar", file.Name())
	require.NoError(t, err)

	_, err = io.Copy(fw, file)
	require.NoError(t, err)
	writer.Close()

	request := httptest.NewRequest("POST", "/testAvatar", body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	ctx := testutil.NewMockContextWith(request)
	ctx = testutil.NewMockContextWithUserID(ctx, testutil.User1.ID)
	testutil.CreateFixtureDb(ctx)

	stg := &mocks.Storage{}
	fileRepo := repository.NewFileRepository()
	domain := NewFileDomain(stg, fileRepo, config.FileConfigs{
		MaxSize: 2,
	})

	stg.On("BulkUpload", mock.Anything, mock.Anything).Return([]*storage.UploadResponse{
		{Url: "28x28.png"},
		{Url: "56x56.png"},
		{Url: "128x128.png"},
	}, nil)
	resp, err := domain.UploadAvatar(ctx, &model.UploadAvatarRequest{})
	require.NoError(t, err)
	var result []*entity.File
	tx := ctx.DB().Model(&entity.File{}).Find(&result)
	require.NoError(t, tx.Error)
	require.Equal(t, 3, len(resp.Urls))
}
