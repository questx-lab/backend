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
	"net/http"

	"github.com/nfnt/resize"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/storage"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func ProcessFormDataImage(
	ctx context.Context, fileStorage storage.Storage, key string,
) (*storage.UploadResponse, error) {
	req := xcontext.HTTPRequest(ctx)
	cfg := xcontext.Configs(ctx).File

	req.Body = http.MaxBytesReader(xcontext.HTTPWriter(ctx), req.Body, cfg.MaxSize)

	if err := req.ParseMultipartForm(cfg.MaxMemory); err != nil {
		xcontext.Logger(ctx).Debugf("Cannot parse multipart form: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Image too large")
	}

	file, header, err := req.FormFile(key)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot get form file: %v", err)
		return nil, errorx.Unknown
	}
	defer file.Close()

	return ResizeImage(ctx, fileStorage, file, header.Filename)
}

func ResizeImage(
	ctx context.Context, fileStorage storage.Storage, file io.ReadSeeker, filename string,
) (*storage.UploadResponse, error) {
	originImg, mime, err := decodeImg(file)
	if err != nil {
		xcontext.Logger(ctx).Warnf("Cannot decode image: %v", err)
		return nil, errorx.New(errorx.BadRequest, "Invalid image")
	}

	cfg := xcontext.Configs(ctx).File
	resizedWidth := cfg.AvatarCropWidth
	resizedHeight := cfg.AvatarCropHeight
	if originImg.Bounds().Dx() > originImg.Bounds().Dy() {
		resizedWidth = 0
	} else {
		resizedHeight = 0
	}

	resizedImg := resize.Resize(resizedWidth, resizedHeight, originImg, resize.Lanczos2)
	subimager, ok := resizedImg.(SubImager)
	if !ok {
		return nil, errorx.New(errorx.Unavailable, "Image doesn't support cropping")
	}

	p1 := image.Point{0, 0}
	p2 := image.Point{resizedImg.Bounds().Dx(), resizedImg.Bounds().Dy()}
	if p2.X > p2.Y {
		p1.X = (p2.X / 2) - int(cfg.AvatarCropWidth)/2
		p2.X = (p2.X / 2) + int(cfg.AvatarCropWidth)/2
	} else {
		p1.Y = (p2.Y / 2) - int(cfg.AvatarCropHeight)/2
		p2.Y = (p2.Y / 2) + int(cfg.AvatarCropHeight)/2
	}

	croppedImg := subimager.SubImage(image.Rectangle{p1, p2})
	b, err := encodeImg(mime, croppedImg)
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot encode image: %v", err)
		return nil, errorx.Unknown
	}

	resp, err := fileStorage.Upload(ctx, &storage.UploadObject{
		Bucket:   string(entity.Image),
		Prefix:   "avatars",
		FileName: fmt.Sprintf("%dx%d-%s", cfg.AvatarCropWidth, cfg.AvatarCropHeight, filename),
		Mime:     mime,
		Data:     b,
	})
	if err != nil {
		xcontext.Logger(ctx).Errorf("Cannot upload image: %v", err)
		return nil, errorx.Unknown
	}

	return resp, nil
}

func decodeImg(data io.ReadSeeker) (image.Image, string, error) {
	fileHeader := make([]byte, 512)
	if _, err := data.Read(fileHeader); err != nil {
		return nil, "", err
	}

	// Set position back to start.
	if _, err := data.Seek(0, 0); err != nil {
		return nil, "", err
	}

	var img image.Image
	var err error
	mime := http.DetectContentType(fileHeader)
	switch mime {
	case "image/jpeg":
		img, err = jpeg.Decode(data)
	case "image/png", "application/octet-stream":
		img, err = png.Decode(data)
	case "image/gif":
		img, err = gif.Decode(data)
	default:
		return nil, "", fmt.Errorf("we just accept jpeg, gif or png")
	}

	return img, mime, err
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
		return nil, fmt.Errorf("we just accept jpeg or png")
	}
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}
