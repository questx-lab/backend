package domain

import (
	"context"
	"io/ioutil"
	"net/http"

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
	cfg := xcontext.Configs(ctx)

	httpReq.Body = http.MaxBytesReader(xcontext.HTTPWriter(ctx), httpReq.Body, cfg.File.MaxSize)

	// max size by MB
	if err := httpReq.ParseMultipartForm(cfg.File.MaxMemory); err != nil {
		xcontext.Logger(ctx).Debugf("Cannot parse multipart form: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Image too large")
	}

	file, header, err := httpReq.FormFile("image")
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get image: %v", err)
		return nil, errorx.Unknown
	}
	defer file.Close()

	name := header.Filename
	mime := header.Header.Values("Content-Type")[0]

	b, err := ioutil.ReadAll(file)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot read image: %v", err)
		return nil, errorx.Unknown
	}

	resp, err := d.storage.Upload(ctx, &storage.UploadObject{
		Bucket:   string(entity.Image),
		Prefix:   "images",
		FileName: name,
		Mime:     mime,
		Data:     b,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot upload image: %v", err)
		return nil, errorx.New(errorx.Internal, "Unable to upload image")
	}

	if err := d.fileRepo.Create(ctx, &entity.File{
		Base:      entity.Base{ID: uuid.NewString()},
		Mime:      mime,
		Name:      resp.FileName,
		Url:       resp.Url,
		CreatedBy: userID,
	}); err != nil {
		xcontext.Logger(ctx).Errorf("Cannot save image in database: %v", err)
		return nil, errorx.Unknown
	}

	return &model.UploadImageResponse{Url: resp.Url}, nil
}
