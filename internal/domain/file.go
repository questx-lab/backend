package domain

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/questx-lab/backend/config"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"

	"github.com/google/uuid"
	"github.com/h2non/bimg"
)

type size struct {
	w int
	h int
}

var (
	AvatarSizes = []size{
		{w: 128, h: 128},
		{w: 56, h: 56},
		{w: 28, h: 28},
	}
)

type FileDomain interface {
	UploadImage(xcontext.Context, *model.UploadImageRequest) (*model.UploadImageResponse, error)
	UploadAvatar(xcontext.Context, *model.UploadAvatarRequest) (*model.UploadAvatarResponse, error)
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

func (d *fileDomain) UploadAvatar(ctx xcontext.Context, req *model.UploadAvatarRequest) (*model.UploadAvatarResponse, error) {
	userID := xcontext.GetRequestUserID(ctx)
	r := ctx.Request()
	// log.Println("UploadAvatar")
	// maximum 2MB
	if err := r.ParseMultipartForm(int64(d.cfg.MaxSize)); err != nil {
		return nil, errorx.New(errorx.BadRequest, "Request must be multipart form")
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {

		return nil, errorx.New(errorx.BadRequest, "Error retrieving the File")
	}
	defer file.Close()

	name := header.Filename
	contentTypes := header.Header.Values("Content-Type")
	if len(contentTypes) == 0 {
		return nil, errorx.New(errorx.BadRequest, "Wrong file content type")
	}
	mime := contentTypes[0]

	b, err := io.ReadAll(file)
	if err != nil {
		return nil, errorx.New(errorx.BadRequest, "Error retrieving the File")
	}

	objs := make([]*storage.UploadObject, len(AvatarSizes))
	for _, size := range AvatarSizes {
		buf, err := bimg.NewImage(b).Resize(size.w, size.h)
		if err != nil {
			return nil, errorx.New(errorx.Internal, "Unable to resize image")
		}

		objs = append(objs, &storage.UploadObject{
			Bucket:   string(entity.Image),
			Prefix:   "avatars",
			FileName: fmt.Sprintf("%dx%d-%s", size.w, size.h, name),
			Mime:     mime,
			Data:     buf,
		})
	}

	uresp, err := d.storage.BulkUpload(ctx, objs)
	if err != nil {
		return nil, errorx.New(errorx.Internal, "Unable to upload image")
	}
	files := make([]*entity.File, len(AvatarSizes))
	urls := make([]string, len(AvatarSizes))
	for _, u := range uresp {
		files = append(files, &entity.File{
			Base: entity.Base{
				ID: uuid.NewString(),
			},
			Mime:      mime,
			Name:      u.FileName,
			Url:       u.Url,
			CreatedBy: userID,
		})
		urls = append(urls, u.Url)
	}
	if err := d.fileRepo.BulkInsert(ctx, files); err != nil {
		return nil, errorx.New(errorx.Internal, "Unable to upload image")
	}

	return &model.UploadAvatarResponse{
		Urls: urls,
	}, nil
}

func (d *fileDomain) UploadImage(ctx xcontext.Context, req *model.UploadImageRequest) (*model.UploadImageResponse, error) {
	userID := xcontext.GetRequestUserID(ctx)
	r := ctx.Request()

	// maximum 2MB
	if err := r.ParseMultipartForm(int64(d.cfg.MaxSize)); err != nil {
		return nil, errorx.New(errorx.BadRequest, "Request must be multipart form")
	}

	file, header, err := r.FormFile("image")
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
		Base: entity.Base{
			ID: uuid.NewString(),
		},
		Mime:      mime,
		Name:      resp.FileName,
		Url:       resp.Url,
		CreatedBy: userID,
	}); err != nil {
		return nil, errorx.New(errorx.Internal, "Unable to upload image")
	}

	return &model.UploadImageResponse{
		Url: resp.Url,
	}, nil
}
