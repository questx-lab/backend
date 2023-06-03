package model

type GetMeRequest struct{}

type GetMeResponse User

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type UpdateUserResponse struct {
	User User `json:"user"`
}

type FollowCommunityRequest struct {
	CommunityHandle string `json:"community_handle"`
	InvitedBy       string `json:"invited_by"`
}

type FollowCommunityResponse struct{}

type GetFollowerRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GetFollowerResponse Follower

type GetAllMyFollowersRequest struct{}

type GetAllMyFollowersResponse struct {
	Followers []Follower `json:"followers"`
}

type GetFollowersRequest struct {
	CommunityHandle string `json:"community_handle"`
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
	UserID          string `json:"user_id"`
	CommunityHandle string `json:"community_handle"`
}

type GetBadgesResponse struct {
	Badges []Badge `json:"badges"`
}

type GetMyBadgesRequest struct {
	CommunityHandle string `json:"community_handle"`
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
