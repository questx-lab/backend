package model

type CreateQuestRequest struct {
	ProjectID      string         `json:"project_id"`
	Type           string         `json:"type"`
	Title          string         `json:"title"`
	Status         string         `json:"status"`
	Description    string         `json:"description"`
	Categories     []string       `json:"categories"`
	Recurrence     string         `json:"recurrence"`
	ValidationData map[string]any `json:"validation_data"`
	Rewards        []Reward       `json:"rewards"`
	ConditionOp    string         `json:"condition_op"`
	Conditions     []Condition    `json:"conditions"`
}

type CreateQuestResponse struct {
	ID string `json:"id"`
}

type GetQuestRequest struct {
	ID                       string `json:"id"`
	IncludeUnclaimableReason bool   `json:"include_unclaimable_reason"`
}

type GetQuestResponse Quest

type GetListQuestRequest struct {
	Q         string `json:"q"`
	ProjectID string `json:"project_id"`
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`

	IncludeUnclaimableReason bool `json:"include_unclaimable_reason"`
}

type GetListQuestResponse struct {
	Quests []Quest `json:"quests,omitempty"`
}

type UpdateQuestRequest struct {
	ID             string         `json:"id"`
	Status         string         `json:"status"`
	Type           string         `json:"type"`
	Title          string         `json:"title,omitempty"`
	Description    string         `json:"description,omitempty"`
	Categories     []string       `json:"categories,omitempty"`
	Recurrence     string         `json:"recurrence,omitempty"`
	ValidationData map[string]any `json:"validation_data,omitempty"`
	Rewards        []Reward       `json:"rewards,omitempty"`
	ConditionOp    string         `json:"condition_op,omitempty"`
	Conditions     []Condition    `json:"conditions,omitempty"`
}

type UpdateQuestResponse struct{}

type DeleteQuestRequest struct {
	ID string `json:"id"`
}

type DeleteQuestResponse struct {
}
