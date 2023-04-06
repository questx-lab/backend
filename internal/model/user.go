package model

type GetUserRequest struct{}

type GetUserResponse struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Name    string `json:"name"`
}

type JoinProjectRequest struct {
	ProjectID string `json:"project_id"`
	InvitedBy string `json:"invite_id"`
}

type JoinProjectResponse struct{}

type GetParticipantRequest struct {
	ProjectID string `json:"project_id"`
}

type GetParticipantResponse struct {
	Points      uint64 `json:"points,omitempty"`
	InviteCode  string `json:"invite_code,omitempty"`
	InvitedBy   string `json:"invited_by,omitempty"`
	InviteCount uint64 `json:"invite_count,omitempty"`
}

type GetInviteRequest struct {
	InviteCode string `json:"invite_code"`
}

type GetInviteResponse struct {
	InvitedBy string  `json:"invited_by,omitempty"`
	Project   Project `json:"project,omitempty"`
}
