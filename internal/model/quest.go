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
	ValidationData string      `json:"validation_data"`
	Awards         []Award     `json:"awards"`
	ConditionOp    string      `json:"condition_op"`
	Conditions     []Condition `json:"conditions"`
}

type CreateQuestResponse struct {
	ID string `json:"id"`
}

type GetShortQuestRequest struct {
	ID string `json:"id"`
}

type GetShortQuestResponse struct {
	ProjectID  string   `json:"project_id"`
	Type       string   `json:"type"`
	Title      string   `json:"title"`
	Categories []string `json:"categories"`
	Recurrence string   `json:"recurrence"`
}
