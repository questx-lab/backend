package model

type UploadImageRequest struct {
	Name string `json:"name"`
}

type UploadImageResponse struct {
	Url string
}

type UploadAvatarRequest struct {
	Name string `json:"name"`
}

type UploadAvatarResponse struct {
	Urls []string
}
