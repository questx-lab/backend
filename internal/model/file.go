package model

type UploadImageRequest struct {
}

type UploadImageResponse struct {
	Url string `json:"url"`
}

type UploadAvatarRequest struct {
}

type UploadAvatarResponse struct {
	Urls []string `json:"urls"`
}
