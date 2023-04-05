package domain

import (
	"github.com/questx-lab/backend/internal/model"
	"github.com/questx-lab/backend/pkg/xcontext"
)

type FileDomain interface {
	UploadImage(xcontext.Context, *model.UploadImageRequest) (*model.UploadImageResponse, error)
}

type fileDomain struct {
}
