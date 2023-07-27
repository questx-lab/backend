package model

type CreateCommunityRequest struct {
	Handle       string `json:"handle"`
	DisplayName  string `json:"display_name"`
	Introduction string `json:"introduction"`
	WebsiteURL   string `json:"website_url"`
	Twitter      string `json:"twitter"`
	ReferralCode string `json:"referral_code"`
	OwnerEmail   string `json:"owner_email"`
}

type CreateCommunityResponse struct {
	Handle string `json:"handle"`
}

type GetCommunitiesRequest struct {
	Q          string `json:"q"`
	ByTrending bool   `json:"by_trending"`
}

type GetCommunitiesResponse struct {
	Communities []Community `json:"communities"`
}

type GetPendingCommunitiesRequest struct{}

type GetPendingCommunitiesResponse struct {
	Communities []Community `json:"communities"`
}

type GetCommunityRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GetCommunityResponse struct {
	Community Community `json:"community"`
}

type GetMyOwnCommunitiesRequest struct{}

type GetMyOwnCommunitiesResponse struct {
	Communities []Community `json:"communities"`
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

type GetMyReferralRequest struct{}

type GetMyReferralResponse struct {
	TotalClaimableCommunities int     `json:"total_claimable_communities"`
	TotalPendingCommunities   int     `json:"total_pending_communities"`
	RewardAmount              float64 `json:"reward_amount"`
}

type GetReferralRequest struct{}

type GetReferralResponse struct {
	Referrals []Referral `json:"referrals"`
}

const (
	ReviewReferralActionApprove = "approve"
	ReviewReferralActionReject  = "reject"
)

type ReviewReferralRequest struct {
	Action          string `json:"action"`
	CommunityHandle string `json:"community_handle"`
}

type ReviewReferralResponse struct{}

type TransferCommunityRequest struct {
	CommunityHandle string `json:"community_handle"`
	ToUserID        string `json:"to_user_id"`
}

type TransferCommunityResponse struct{}

type ApprovePendingCommunityRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type ApprovePendingCommunityResponse struct{}

type GetDiscordRoleRequest struct {
	CommunityHandle string `json:"community_handle"`
	IncludeAll      bool   `json:"include_all"`
}

type GetDiscordRoleResponse struct {
	Roles []DiscordRole `json:"roles"`
}

type AssignRoleRequest struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

type AssignRoleResponse struct {
}

type DeleteUserCommunityRoleRequest struct {
	UserID          string   `json:"user_id"`
	CommunityHandle string   `json:"community_handle"`
	RoleIDs         []string `json:"role_ids"`
}

type DeleteUserCommunityRoleResponse struct {
}
