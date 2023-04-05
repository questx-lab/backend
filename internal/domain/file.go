package domain

import (
	"encoding/base64"
	"strings"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/google/uuid"
)

type FileDomain interface {
	UploadImage(xcontext.Context, *model.UploadImageRequest) (*model.UploadImageResponse, error)
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

func (d *fileDomain) UploadImage(ctx xcontext.Context, req *model.UploadImageRequest) (*model.UploadImageResponse, error) {
	userID := ctx.GetUserID()
	b, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		return nil, errorx.New(errorx.BadRequest, "Data must be base64")
	}
	num := strings.Index(req.Data, ",")
	if num < 0 {
		return nil, errorx.New(errorx.BadRequest, "Data must be base64")
	}
	mime := req.Data[5 : num-7]
	if mime != "image/jpeg" && mime != "image/png" {
		errorx.New(errorx.BadRequest, "Data must be image")
	}
	resp, err := d.storage.Upload(ctx, &storage.UploadObject{
		Bucket:   string(entity.Image),
		Prefix:   "images",
		FileName: req.Name,
		Mime:     mime,
		Data:     b,
	})
	if err != nil {
		return nil, errorx.New(errorx.Internal, "Unable to upload image")
	}

	if err := d.fileRepo.Create(ctx, &entity.File{
		Base: entity.Base{
			ID: uuid.NewString(),
		},
		Mine:      mime,
		Name:      req.Name,
		Url:       resp.Url,
		CreatedBy: userID,
	}); err != nil {
		return nil, errorx.New(errorx.Internal, "Unable to upload image")
	}
	return &model.UploadImageResponse{
		Url: resp.Url,
	}, nil
}
