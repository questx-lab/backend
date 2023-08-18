package model

type CreateNFTRequest struct {
	ID              int64  `json:"id"`
	CommunityHandle string `json:"community_handle"`
	Title           string `json:"title"`
	ImageUrl        string `json:"image_url"`
	Amount          int    `json:"amount"`
	Description     string `json:"description"`
	Chain           string `json:"chain"`
}

type CreateNFTResponse struct{}

type GetNFTRequest struct {
	NftID int64 `json:"nft_id"`
}

type GetNFTResponse struct {
	NFT NonFungibleToken `json:"nft"`
}

type GetNFTsRequest struct {
	NftIDs []int64 `json:"nft_ids"`
}

type GetNFTsResponse struct {
	NFTs []NonFungibleToken `json:"nfts"`
}

type GetNFTsByCommunityRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GetNFTsByCommunityResponse struct {
	NFTs []NonFungibleToken `json:"nfts"`
}
