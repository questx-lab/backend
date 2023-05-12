package model

type GenerateAPIKeyRequest struct {
	ProjectID string `json:"project_id"`
}

type GenerateAPIKeyResponse struct {
	Key string `json:"key"`
}

type RegenerateAPIKeyRequest struct {
	ProjectID string `json:"project_id"`
}

type RegenerateAPIKeyResponse struct {
	Key string `json:"key"`
}

type RevokeAPIKeyRequest struct {
	ProjectID string `json:"project_id"`
}

type RevokeAPIKeyResponse struct{}
