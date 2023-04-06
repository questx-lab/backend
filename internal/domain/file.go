package domain

import (
	"io/ioutil"
	"net/http"

	"github.com/questx-lab/backend/config"
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
	cfg      config.FileConfigs
}

func NewFileDomain(
	storage storage.Storage,
	fileRepo repository.FileRepository,
	cfg config.FileConfigs,
) FileDomain {
	return &fileDomain{
		storage:  storage,
		fileRepo: fileRepo,
		cfg:      cfg,
	}
}

func (d *fileDomain) UploadImage(ctx xcontext.Context, req *model.UploadImageRequest) (*model.UploadImageResponse, error) {
	userID := xcontext.GetRequestUserID(ctx)
	r := ctx.Request()

	// maximum 2MB
	if err := r.ParseMultipartForm(10 << 2); err != nil {
		return nil, errorx.New(errorx.BadRequest, "Request must be multipart form")
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		return nil, errorx.New(errorx.BadRequest, "Error retrieving the File")
	}
	defer file.Close()

	fileHeader := make([]byte, 512)
	// Copy the headers into the FileHeader buffer
	if _, err := file.Read(fileHeader); err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errorx.New(errorx.BadRequest, "Error retrieving the File")
	}
	mime := http.DetectContentType(fileHeader)

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
