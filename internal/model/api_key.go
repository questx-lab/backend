package model

type GenerateAPIKeyRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GenerateAPIKeyResponse struct {
	Key string `json:"key"`
}

type RegenerateAPIKeyRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type RegenerateAPIKeyResponse struct {
	Key string `json:"key"`
}

type RevokeAPIKeyRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type RevokeAPIKeyResponse struct{}
