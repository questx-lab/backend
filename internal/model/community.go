package model

type CreateCommunityRequest struct {
	Handle             string   `json:"handle"`
	DisplayName        string   `json:"display_name"`
	Introduction       string   `json:"introduction"`
	WebsiteURL         string   `json:"website_url"`
	DevelopmentStage   string   `json:"development_stage"`
	TeamSize           int      `json:"team_size"`
	SharedContentTypes []string `json:"shared_content_types"`
	Twitter            string   `json:"twitter"`
	ReferralCode       string   `json:"referral_code"`
}

type CreateCommunityResponse struct {
	ID string `json:"id"`
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
	ID string `json:"id"`
}

type GetCommunityResponse struct {
	Community Community `json:"community"`
}

type UpdateCommunityRequest struct {
	ID                 string   `json:"id"`
	DisplayName        string   `json:"display_name"`
	Introduction       string   `json:"introduction"`
	WebsiteURL         string   `json:"website_url"`
	DevelopmentStage   string   `json:"development_stage"`
	TeamSize           int      `json:"team_size"`
	SharedContentTypes []string `json:"shared_content_types"`
	Twitter            string   `json:"twitter"`
}

type UpdateCommunityResponse struct{}

type UpdateCommunityDiscordRequest struct {
	ID          string `json:"id"`
	ServerID    string `json:"server_id"`
	AccessToken string `json:"access_token"`
}

type UpdateCommunityDiscordResponse struct{}

type DeleteCommunityRequest struct {
	ID string `json:"id"`
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

type GetPendingReferralRequest struct{}

type GetPendingReferralResponse struct {
	Communities []Community `json:"communities"`
}

type ApproveReferralRequest struct {
	CommunityIDs []string `json:"community_ids"`
}

type ApproveReferralResponse struct{}
