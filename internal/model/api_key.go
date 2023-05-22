package model

type GenerateAPIKeyRequest struct {
	CommunityID string `json:"community_id"`
}

type GenerateAPIKeyResponse struct {
	Key string `json:"key"`
}

type RegenerateAPIKeyRequest struct {
	CommunityID string `json:"community_id"`
}

type RegenerateAPIKeyResponse struct {
	Key string `json:"key"`
}

type RevokeAPIKeyRequest struct {
	CommunityID string `json:"community_id"`
}

type RevokeAPIKeyResponse struct{}
