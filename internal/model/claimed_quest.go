package model

type ClaimQuestRequest struct {
	QuestID        string `json:"quest_id"`
	SubmissionData string `json:"submission_data"`
	WalletAddress  string `json:"wallet_address"`
}

type ClaimQuestResponse struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ClaimReferralRequest struct {
	WalletAddress string `json:"wallet_address"`
}

type ClaimReferralResponse struct{}

type GetClaimedQuestRequest struct {
	ID string `json:"id"`
}

type GetClaimedQuestResponse ClaimedQuest

type GetListClaimedQuestRequest struct {
	CommunityHandle string `json:"community_handle"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`

	QuestID    string `json:"quest_id"`
	UserID     string `json:"user_id"`
	Recurrence string `json:"recurrence"`
	Status     string `json:"status"`

	ReverseTime bool `json:"reverse_time"`
}

type GetListClaimedQuestResponse struct {
	ClaimedQuests []ClaimedQuest `json:"claimed_quests"`
}

type ReviewRequest struct {
	Action  string   `json:"action"`
	Comment string   `json:"comment"`
	IDs     []string `json:"ids"`
}

type ReviewResponse struct{}

type ReviewAllRequest struct {
	Action          string `json:"action"`
	Comment         string `json:"comment"`
	CommunityHandle string `json:"community_handle"`

	QuestIDs    []string `json:"quest_ids"`
	UserIDs     []string `json:"user_ids"`
	Recurrences []string `json:"recurrences"`
	Statuses    []string `json:"statuses"`
	Excludes    []string `json:"excludes"`
}

type ReviewAllResponse struct {
	Quantity int `json:"quantity"`
}

type GivePointRequest struct {
	CommunityHandle string `json:"community_handle"`
	UserID          string `json:"user_id"`
	Points          uint64 `json:"points"`
}

type GivePointResponse struct{}
