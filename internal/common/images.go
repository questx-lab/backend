package common

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/nfnt/resize"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type size struct {
	w int
	h int
}

func (s size) String() string {
	return fmt.Sprintf("%dx%d", s.w, s.h)
}

var (
	AvatarSizes = []size{
		{w: 512, h: 512},
		{w: 128, h: 128},
		{w: 32, h: 32},
	}
)

func ProcessImage(ctx context.Context, fileStorage storage.Storage, key string) ([]*storage.UploadResponse, error) {
	req := xcontext.HTTPRequest(ctx)
	if err := req.ParseMultipartForm(xcontext.Configs(ctx).File.MaxSize); err != nil {
		return nil, errorx.New(errorx.BadRequest, "Request must be multipart form")
	}

	file, header, err := req.FormFile(key)
	if err != nil {
		return nil, errorx.New(errorx.BadRequest, "Error retrieving the File")
	}
	defer file.Close()

	mime := header.Header.Get("Content-Type")
	img, err := decodeImg(mime, file)
	if err != nil {
		return nil, err
	}

	objs := make([]*storage.UploadObject, 0, len(AvatarSizes))
	for _, size := range AvatarSizes {
		img := resize.Resize(uint(size.w), uint(size.h), img, resize.Lanczos2)
		b, err := encodeImg(mime, img)
		if err != nil {
			xcontext.Logger(ctx).Errorf("Cannot encode image: %v", err)
			return nil, errorx.Unknown
		}

		objs = append(objs, &storage.UploadObject{
			Bucket:   string(entity.Image),
			Prefix:   "avatars",
			FileName: fmt.Sprintf("%dx%d-%s", size.w, size.h, header.Filename),
			Mime:     mime,
			Data:     b,
		})
	}

	uresp, err := fileStorage.BulkUpload(ctx, objs)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot upload image: %v", err)
		return nil, errorx.Unknown
	}

	return uresp, nil
}

func decodeImg(mime string, data io.Reader) (img image.Image, err error) {
	switch mime {
	case "image/jpeg":
		img, err = jpeg.Decode(data)
	case "image/png", "application/octet-stream":
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
	case "image/png", "application/octet-stream":
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
