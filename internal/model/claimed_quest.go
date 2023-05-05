package model

type ClaimedQuest struct {
	ID         string `json:"id,omitempty"`
	QuestID    string `json:"quest_id,omitempty"`
	Quest      Quest  `json:"quest,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	User       User   `json:"user,omitempty"`
	Status     string `json:"status,omitempty"`
	Input      string `json:"input,omitempty"`
	ReviewerID string `json:"reviewer_id,omitempty"`
	ReviewerAt string `json:"reviewer_at,omitempty"`
	Comment    string `json:"comment,omitempty"`
}

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

type GetClaimedQuestResponse ClaimedQuest

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

type ReviewRequest struct {
	Action string   `json:"action"`
	IDs    []string `json:"ids"`
}

type ReviewResponse struct{}

type ReviewAllRequest struct {
	Action        string   `json:"action"`
	ProjectID     string   `json:"project_id"`
	FilterQuestID string   `json:"filter_quest_id"`
	FilterUserID  string   `json:"filter_user_id"`
	Excludes      []string `json:"excludes"`
}

type ReviewAllResponse struct {
	Quantity int `json:"quantity,omitempty"`
}

type GiveRewardRequest struct {
	ProjectID string         `json:"project_id"`
	UserID    string         `json:"user_id"`
	Type      string         `json:"type"`
	Data      map[string]any `json:"data"`
}

type GiveRewardResponse struct{}
