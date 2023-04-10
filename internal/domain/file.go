package domain

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
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
	"github.com/nfnt/resize"
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
	storage       storage.Storage
	fileRepo      repository.FileRepository
	cfg           config.FileConfigs
	MaxUploadSize int
}

func NewFileDomain(
	storage storage.Storage,
	fileRepo repository.FileRepository,
	cfg config.FileConfigs,
) FileDomain {

	return &fileDomain{
		storage:       storage,
		fileRepo:      fileRepo,
		cfg:           cfg,
		MaxUploadSize: cfg.MaxSize * 1024 * 1024,
	}
}

func decodeImg(mime string, data io.Reader) (img image.Image, err error) {
	switch mime {
	case "image/jpeg":
		img, err = jpeg.Decode(data)
	case "image/png":
		img, err = png.Decode(data)
	case "image/gif":
		img, err = gif.Decode(data)
	default:
		return nil, fmt.Errorf("We just accept jpeg, gif or png")
	}
	return img, err
}

func encodeImg(mime string, img image.Image) (b []byte, err error) {
	buf := new(bytes.Buffer)

	switch mime {
	case "image/jpeg":
		err = jpeg.Encode(buf, img, nil)
	case "image/png":
		err = jpeg.Encode(buf, img, nil)
	case "image/gif":
		err = gif.Encode(buf, img, nil)
	default:
		return nil, fmt.Errorf("We just accept jpeg or png")
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (d *fileDomain) UploadAvatar(ctx xcontext.Context, req *model.UploadAvatarRequest) (*model.UploadAvatarResponse, error) {
	userID := xcontext.GetRequestUserID(ctx)
	r := ctx.Request()

	// max size by MB
	if err := r.ParseMultipartForm(int64(d.MaxUploadSize)); err != nil {
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

	img, err := decodeImg(mime, file)
	if err != nil {
		return nil, err
	}
	objs := make([]*storage.UploadObject, 0, len(AvatarSizes))
	for _, size := range AvatarSizes {
		img := resize.Resize(uint(size.w), uint(size.h), img, resize.Lanczos2)
		b, err := encodeImg(mime, img)
		if err != nil {
			return nil, err
		}

		objs = append(objs, &storage.UploadObject{
			Bucket:   string(entity.Image),
			Prefix:   "avatars",
			FileName: fmt.Sprintf("%dx%d-%s", size.w, size.h, name),
			Mime:     mime,
			Data:     b,
		})
	}

	uresp, err := d.storage.BulkUpload(ctx, objs)
	if err != nil {
		return nil, errorx.New(errorx.Internal, "Unable to upload image")
	}
	files := make([]*entity.File, 0, len(AvatarSizes))
	urls := make([]string, 0, len(AvatarSizes))
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

	// max size by MB
	if err := r.ParseMultipartForm(int64(d.MaxUploadSize)); err != nil {
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
