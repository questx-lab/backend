package model

type GetUserRequest struct{}

type GetUserResponse struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Name    string `json:"name"`
}

type JoinProjectRequest struct {
	ProjectID  string `json:"project_id"`
	ReferralID string `json:"invite_id"`
}

type JoinProjectResponse struct{}

type GetParticipantRequest struct {
	ProjectID string `json:"project_id"`
}

type GetParticipantResponse struct {
	Points        uint64 `json:"points,omitempty"`
	ReferralCode  string `json:"referral_code,omitempty"`
	ReferralID    string `json:"invited_by,omitempty"`
	ReferralCount uint64 `json:"invite_count,omitempty"`
}

type GetReferralInfoRequest struct {
	ReferralCode string `json:"referral_code"`
}

type GetReferralInfoResponse struct {
	ReferralID string  `json:"referral_id,omitempty"`
	Project    Project `json:"project,omitempty"`
}
