package model

type UploadImageRequest struct {
	Name string `json:"name"`
	Mime string `json:"mime"`
	Data string `json:"data"`
}

type UploadImageResponse struct {
	Url string
}
