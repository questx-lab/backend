package model

type GetUserRequest struct{}

type GetUserResponse User

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type UpdateUserResponse struct{}

type FollowProjectRequest struct {
	ProjectID string `json:"project_id"`
	InvitedBy string `json:"invite_id"`
}

type FollowProjectResponse struct{}

type GetParticipantRequest struct {
	ProjectID string `json:"project_id"`
}

type GetParticipantResponse Participant

type GetListParticipantRequest struct {
	ProjectID string `json:"project_id"`
}

type GetListParticipantResponse struct {
	Participants []Participant
}

type GetInviteRequest struct {
	InviteCode string `json:"invite_code"`
}

type GetInviteResponse struct {
	User    User    `json:"user"`
	Project Project `json:"project"`
}

type GetBadgesRequest struct {
	UserID    string `json:"user_id"`
	ProjectID string `json:"project_id"`
}

type GetBadgesResponse struct {
	Badges []Badge `json:"badges"`
}

type GetMyBadgesRequest struct {
	ProjectID string `json:"project_id"`
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
