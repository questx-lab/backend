package domain

import (
	"context"
	"io/ioutil"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/google/uuid"
)

type FileDomain interface {
	UploadImage(context.Context, *model.UploadImageRequest) (*model.UploadImageResponse, error)
}

type fileDomain struct {
	storage  storage.Storage
	fileRepo repository.FileRepository
}

func NewFileDomain(
	storage storage.Storage,
	fileRepo repository.FileRepository,
) FileDomain {

	return &fileDomain{
		storage:  storage,
		fileRepo: fileRepo,
	}
}

func (d *fileDomain) UploadImage(ctx context.Context, req *model.UploadImageRequest) (*model.UploadImageResponse, error) {
	userID := xcontext.RequestUserID(ctx)
	httpReq := xcontext.HTTPRequest(ctx)

	// max size by MB
	if err := httpReq.ParseMultipartForm(xcontext.Configs(ctx).File.MaxSize); err != nil {
		return nil, errorx.New(errorx.BadRequest, "Request must be multipart form")
	}

	file, header, err := httpReq.FormFile("image")
	if err != nil {
		return nil, errorx.New(errorx.BadRequest, "Error retrieving the File")
	}
	defer file.Close()

	name := header.Filename
	mime := header.Header.Values("Content-Type")[0]

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errorx.New(errorx.BadRequest, "Error retrieving the File")
	}

	resp, err := d.storage.Upload(ctx, &storage.UploadObject{
		Bucket:   string(entity.Image),
		Prefix:   "images",
		FileName: name,
		Mime:     mime,
		Data:     b,
	})
	if err != nil {
		return nil, errorx.New(errorx.Internal, "Unable to upload image")
	}

	if err := d.fileRepo.Create(ctx, &entity.File{
		Base:      entity.Base{ID: uuid.NewString()},
		Mime:      mime,
		Name:      resp.FileName,
		Url:       resp.Url,
		CreatedBy: userID,
	}); err != nil {
		return nil, errorx.New(errorx.Internal, "Unable to upload image")
	}

	return &model.UploadImageResponse{Url: resp.Url}, nil
}
