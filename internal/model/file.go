package model

type UploadImageRequest struct {
}

type UploadImageResponse struct {
	Url string
}

type UploadAvatarRequest struct {
}

type UploadAvatarResponse struct {
	Urls []string
}
