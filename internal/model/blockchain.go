package model

type CreateBlockchainRequest struct {
	Chain                string `json:"chain"`
	ChainID              int64  `json:"chain_id"`
	UseExternalRPC       bool   `json:"use_external_rpc"`
	UseEip1559           bool   `json:"use_eip_1559"`
	BlockTime            int    `json:"block_time"`
	AdjustTime           int    `json:"adjust_time"`
	ThresholdUpdateBlock int    `json:"threshold_update_block"`
}

type CreateBlockchainResponse struct{}

type CreateBlockchainConnectionRequest struct {
	Chain string `json:"chain"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}

type CreateBlockchainConnectionResponse struct{}

type DeleteBlockchainConnectionRequest struct {
	Chain string `json:"chain"`
	URL   string `json:"url"`
}

type DeleteBlockchainConnectionResponse struct{}

type GetCommunityWalletAddressRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GetCommunityWalletAddressResponse struct {
	WalletAddress string `json:"wallet_address"`
}

type CreateBlockchainTokenRequest struct {
	Chain   string `json:"chain"`
	Address string `json:"address"`
}

type CreateBlockchainTokenResponse struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Decimals int    `json:"decimals"`
}
