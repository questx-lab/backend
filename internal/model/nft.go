package model

type CreateNftsRequest struct {
	ID              int64  `json:"id"`
	CommunityHandle string `json:"community_handle"`
	Title           string `json:"title"`
	ImageUrl        string `json:"image_url"`
	Amount          int64  `json:"amount"`
	Description     string `json:"description"`
	Chain           string `json:"chain"`
}

type CreateNftsResponse struct{}

type GetNftsRequest struct {
	SetID string `json:"set_id"`
}

type GetNftsResponse struct{}
