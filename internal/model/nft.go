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
	NFT NFT `json:"nft"`
}

type GetNFTsRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GetNFTsResponse struct {
	NFTs []NFT `json:"nfts"`
}
