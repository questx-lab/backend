package model

type ClaimQuestRequest struct {
	QuestID string `json:"quest_id"`
	Input   string `json:"input"`
}

type ClaimQuestResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type GetClaimedQuestRequest struct {
	ID string `json:"id"`
}

type GetClaimedQuestResponse struct {
	QuestID    string `json:"quest_id,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	Input      string `json:"input,omitempty"`
	Status     string `json:"status,omitempty"`
	ReviewerID string `json:"reviewer_id,omitempty"`
	ReviewerAt string `json:"reviewer_at,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

type ClaimedQuest struct {
	QuestID    string `json:"quest_id,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	Status     string `json:"status,omitempty"`
	ReviewerID string `json:"reviewer_id,omitempty"`
	ReviewerAt string `json:"reviewer_at,omitempty"`
}

type GetListClaimedQuestRequest struct {
	ProjectID string `json:"project_id"`

	Offset int `json:"offset"`
	Limit  int `json:"limit"`

	FilterQuestID string `json:"filter_quest_id"`
	FilterUserID  string `json:"filter_user_id"`
	FilterStatus  string `json:"filter_status"`
}

type GetListClaimedQuestResponse struct {
	ClaimedQuests []ClaimedQuest `json:"claimed_quests"`
}

type ReviewClaimedQuestRequest struct {
	ID     string `json:"id"`
	Action string `json:"action"`
}

type ReviewClaimedQuestResponse struct{}

type GiveRewardRequest struct {
	ProjectID string         `json:"project_id"`
	UserID    string         `json:"user_id"`
	Type      string         `json:"type"`
	Data      map[string]any `json:"data"`
}

type GiveRewardResponse struct{}
