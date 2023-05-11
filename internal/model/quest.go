package model

type Award struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Condition struct {
	Type  string `json:"type"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

type CreateQuestRequest struct {
	ProjectID      string      `json:"project_id"`
	Type           string      `json:"type"`
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	Categories     []string    `json:"categories"`
	Recurrence     string      `json:"recurrence"`
	ValidationData any         `json:"validation_data"`
	Awards         []Award     `json:"awards"`
	ConditionOp    string      `json:"condition_op"`
	Conditions     []Condition `json:"conditions"`
}

type CreateQuestResponse struct {
	ID string `json:"id"`
}

type GetQuestRequest struct {
	ID string `json:"id"`
}

type GetQuestResponse struct {
	ProjectID      string      `json:"project_id,omitempty"`
	Type           string      `json:"type,omitempty"`
	Status         string      `json:"status,omitempty"`
	Title          string      `json:"title,omitempty"`
	Description    string      `json:"description,omitempty"`
	Categories     []string    `json:"categories,omitempty"`
	Recurrence     string      `json:"recurrence,omitempty"`
	ValidationData any         `json:"validation_data,omitempty"`
	Awards         []Award     `json:"awards,omitempty"`
	ConditionOp    string      `json:"condition_op,omitempty"`
	Conditions     []Condition `json:"conditions,omitempty"`
	CreatedAt      string      `json:"created_at,omitempty"`
	UpdatedAt      string      `json:"updated_at,omitempty"`
}

type GetListQuestRequest struct {
	ProjectID string `json:"project_id"`
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
}

type ShortQuest struct {
	ID         string   `json:"id,omitempty"`
	Type       string   `json:"type,omitempty"`
	Title      string   `json:"title,omitempty"`
	Status     string   `json:"status,omitempty"`
	Categories []string `json:"categories,omitempty"`
	Recurrence string   `json:"recurrence,omitempty"`
}

type GetListQuestResponse struct {
	Quests []ShortQuest `json:"quests,omitempty"`
}

type UpdateQuestRequest struct {
	ID             string         `json:"id"`
	Status         string         `json:"status"`
	Type           string         `json:"type"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Categories     []string       `json:"categories"`
	Recurrence     string         `json:"recurrence"`
	ValidationData map[string]any `json:"validation_data"`
	Rewards        []Reward       `json:"rewards"`
	ConditionOp    string         `json:"condition_op"`
	Conditions     []Condition    `json:"conditions"`
}

type UpdateQuestResponse struct{}

type DeleteQuestRequest struct {
	ID string `json:"id"`
}

type DeleteQuestResponse struct {
}
