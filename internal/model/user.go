package model

type GetUserRequest struct{}

type GetUserResponse struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Name    string `json:"name"`
}

type FollowProjectRequest struct {
	ProjectID string `json:"project_id"`
	InvitedBy string `json:"invite_id"`
}

type FollowProjectResponse struct{}

type Participant struct {
	UserID      string `json:"user_id,omitempty"`
	Points      uint64 `json:"points,omitempty"`
	InviteCode  string `json:"invite_code,omitempty"`
	InvitedBy   string `json:"invited_by,omitempty"`
	InviteCount uint64 `json:"invite_count,omitempty"`
}

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
	InvitedBy string  `json:"invited_by,omitempty"`
	Project   Project `json:"project,omitempty"`
}
