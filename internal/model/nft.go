package model

type CreateNFTRequest struct {
	ID              int64  `json:"id"`
	CommunityHandle string `json:"community_handle"`
	Title           string `json:"title"`
	ImageUrl        string `json:"image_url"`
	Amount          int64  `json:"amount"`
	Description     string `json:"description"`
	Chain           string `json:"chain"`
}

type CreateNFTResponse struct{}

type GetNFTRequest struct {
	ID int64 `json:"id"`
}

type GetNFTResponse struct {
	Title       string `json:"title"`
	ImageUrl    string `json:"image_url"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
	Chain       string `json:"chain"`
	CreatedBy   string `json:"created_by"`

	PendingAmount int64 `json:"pending_amount"`
	ActiveAmount  int64 `json:"active_amount"`
	FailureAmount int64 `json:"failure_amount"`
}
