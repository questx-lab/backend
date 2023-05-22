package model

type GetUserRequest struct{}

type GetUserResponse User

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type UpdateUserResponse struct{}

type FollowCommunityRequest struct {
	CommunityID string `json:"community_id"`
	InvitedBy   string `json:"invited_by"`
}

type FollowCommunityResponse struct{}

type GetFollowerRequest struct {
	CommunityID string `json:"community_id"`
}

type GetFollowerResponse Follower

type GetFollowersRequest struct {
	CommunityID string `json:"community_id"`
}

type GetFollowersResponse struct {
	Followers []Follower `json:"followers"`
}

type GetInviteRequest struct {
	InviteCode string `json:"invite_code"`
}

type GetInviteResponse struct {
	User      User      `json:"user"`
	Community Community `json:"community"`
}

type GetBadgesRequest struct {
	UserID      string `json:"user_id"`
	CommunityID string `json:"community_id"`
}

type GetBadgesResponse struct {
	Badges []Badge `json:"badges"`
}

type GetMyBadgesRequest struct {
	CommunityID string `json:"community_id"`
}

type GetMyBadgesResponse struct {
	Badges []Badge `json:"badges"`
}

type AssignGlobalRoleRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type AssignGlobalRoleResponse struct{}

type UploadAvatarRequest struct {
	// Avatar data is included in form-data.
}

type UploadAvatarResponse struct{}
