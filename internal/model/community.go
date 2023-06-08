package model

type CreateCommunityRequest struct {
	Handle       string `json:"handle"`
	DisplayName  string `json:"display_name"`
	Introduction string `json:"introduction"`
	WebsiteURL   string `json:"website_url"`
	Twitter      string `json:"twitter"`
	InviteCode   string `json:"invite_code"`
}

type CreateCommunityResponse struct {
	Handle string `json:"handle"`
}

type GetCommunitiesRequest struct {
	Q          string `json:"q"`
	Offset     int    `json:"offset"`
	Limit      int    `json:"limit"`
	ByTrending bool   `json:"by_trending"`
}

type GetCommunitiesResponse struct {
	Communities []Community `json:"communities"`
}

type GetCommunityRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GetCommunityResponse struct {
	Community Community `json:"community"`
}

type UpdateCommunityRequest struct {
	CommunityHandle string `json:"community_handle"`
	DisplayName     string `json:"display_name"`
	Introduction    string `json:"introduction"`
	WebsiteURL      string `json:"website_url"`
	Twitter         string `json:"twitter"`
}

type UpdateCommunityResponse struct {
	Community Community `json:"community"`
}

type UpdateCommunityDiscordRequest struct {
	CommunityHandle string `json:"community_handle"`
	ServerID        string `json:"server_id"`

	// For Authorization Code flow.
	AccessToken string `json:"access_token"`

	// For Authorization Code with PKCE flow.
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	RedirectURI  string `json:"redirect_uri"`

	// For Implicit flow.
	IDToken string `json:"id_token"`
}

type UpdateCommunityDiscordResponse struct{}

type DeleteCommunityRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type DeleteCommunityResponse struct{}

type GetFollowingCommunitiesRequest struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type GetFollowingCommunitiesResponse struct {
	Communities []Community `json:"communities"`
}

type UploadCommunityLogoRequest struct {
	// Logo data is included in form-data.
}

type UploadCommunityLogoResponse struct{}

type GetMyInvitedCommunitiesRequest struct{}

type GetInvitedCommunitiesResponse struct {
	TotalClaimableCommunities int     `json:"total_claimable_communities"`
	TotalPendingCommunities   int     `json:"total_pending_communities"`
	RewardAmount              float64 `json:"reward_amount"`
}

type GetPendingInviteCommunitiesRequest struct{}

type GetPendingInviteCommunitiesResponse struct {
	Communities []Community `json:"communities"`
}

type ApproveInvitedCommunitiesRequest struct {
	CommunityHandles []string `json:"community_handles"`
}

type ApproveInvitedCommunitiesResponse struct{}

type TransferCommunityRequest struct {
	CommunityHandle string `json:"community_handle"`
	ToID            string `json:"to_id"`
}

type TransferCommunityResponse struct{}
