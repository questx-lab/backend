package model

type GetAllBadgeNamesRequest struct{}

type GetAllBadgeNamesResponse struct {
	Names []string `json:"names"`
}

type GetAllBadgesRequest struct{}

type GetAllBadgesResponse struct {
	Badges []Badge `json:"badges"`
}

type UpdateBadgeRequest struct {
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Value       int    `json:"value"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
}

type UpdateBadgeResponse struct{}

type GetUserBadgeDetailsRequest struct {
	UserID          string `json:"user_id"`
	CommunityHandle string `json:"community_handle"`
}

type GetUserBadgeDetailsResponse struct {
	BadgeDetails []BadgeDetail `json:"badge_details"`
}

type GetMyBadgeDetailsRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GetMyBadgeDetailsResponse struct {
	BadgeDetails []BadgeDetail `json:"badge_details"`
}
