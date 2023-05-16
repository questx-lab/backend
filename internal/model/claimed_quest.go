package model

type ClaimQuestRequest struct {
	QuestID string `json:"quest_id"`
	Input   string `json:"input"`
}

type ClaimQuestResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type ClaimReferralRequest struct {
	ProjectID string `json:"project_id"`
}

type ClaimReferralResponse struct{}

type GetClaimedQuestRequest struct {
	ID string `json:"id"`
}

type GetClaimedQuestResponse ClaimedQuest

type GetListClaimedQuestRequest struct {
	ProjectID string `json:"project_id"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`

	QuestID    string `json:"quest_id"`
	UserID     string `json:"user_id"`
	Recurrence string `json:"recurrence"`
	Status     string `json:"status"`
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
	Action    string `json:"action"`
	Comment   string `json:"comment"`
	ProjectID string `json:"project_id"`

	QuestIDs    []string `json:"quest_ids"`
	UserIDs     []string `json:"user_ids"`
	Recurrences []string `json:"recurrences"`
	Excludes    []string `json:"excludes"`
}

type ReviewAllResponse struct {
	Quantity int `json:"quantity"`
}

type GiveRewardRequest struct {
	ProjectID string         `json:"project_id"`
	UserID    string         `json:"user_id"`
	Type      string         `json:"type"`
	Data      map[string]any `json:"data"`
}

type GiveRewardResponse struct{}
