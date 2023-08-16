package model

type CreateNftsRequest struct {
	CommunityHandle string `json:"community_handle"`
	Title           string `json:"title"`
	ImageUrl        string `json:"image_url"`
	Chain           string `json:"chain"`
	Amount          int64  `json:"amount"`
	Description     string `json:"description"`
}

type CreateNftsResponse struct{}

type GetNftsRequest struct {
	SetID string `json:"set_id"`
}

type GetNftsResponse struct{}
