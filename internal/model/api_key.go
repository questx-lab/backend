package model

type GenerateAPIKeyRequest struct {
	ProjectID string `json:"project_id"`
}

type GenerateAPIKeyResponse struct {
	Key string `json:"key,omitempty"`
}

type RevokeAPIKeyRequest struct {
	ProjectID string `json:"project_id"`
}

type RevokeAPIKeyResponse struct{}
