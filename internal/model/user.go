package model

type GetMeRequest struct{}

type GetMeResponse struct {
	User User `json:"user"`
}

type GetUserRequest struct {
	UserID string `json:"user_id"`
}

type GetUserResponse struct {
	User User `json:"user"`
}

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

type UnFollowCommunityRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type UnFollowCommunityResponse struct{}

type GetInviteRequest struct {
	InviteCode string `json:"invite_code"`
}

type GetInviteResponse struct {
	User      User      `json:"user"`
	Community Community `json:"community"`
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

type CountTotalUsersRequest struct{}

type CountTotalUsersResponse struct {
	Total int `json:"total"`
}
