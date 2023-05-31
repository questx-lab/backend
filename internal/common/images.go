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

func ProcessImage(ctx context.Context, fileStorage storage.Storage, key string) (*storage.UploadResponse, error) {
	req := xcontext.HTTPRequest(ctx)
	cfg := xcontext.Configs(ctx).File
	if err := req.ParseMultipartForm(cfg.MaxSize); err != nil {
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

	resizedImg := resize.Resize(cfg.AvatarCropWidth, cfg.AvatarCropHeight, img, resize.Lanczos2)
	b, err := encodeImg(mime, resizedImg)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot encode image: %v", err)
		return nil, errorx.Unknown
	}

	resp, err := fileStorage.Upload(ctx, &storage.UploadObject{
		Bucket:   string(entity.Image),
		Prefix:   "avatars",
		FileName: fmt.Sprintf("%dx%d-%s", cfg.AvatarCropWidth, cfg.AvatarCropHeight, header.Filename),
		Mime:     mime,
		Data:     b,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot upload image: %v", err)
		return nil, errorx.Unknown
	}

	return resp, nil
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
