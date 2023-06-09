package model

type CreateQuestRequest struct {
	CommunityHandle string         `json:"community_handle"`
	Type            string         `json:"type"`
	Title           string         `json:"title"`
	Status          string         `json:"status"`
	Description     string         `json:"description"`
	CategoryID      string         `json:"category_id"`
	Recurrence      string         `json:"recurrence"`
	ValidationData  map[string]any `json:"validation_data"`
	Points          uint64         `json:"points"`
	Rewards         []Reward       `json:"rewards"`
	ConditionOp     string         `json:"condition_op"`
	Conditions      []Condition    `json:"conditions"`
	IsHighlight     bool           `json:"is_highlight"`
}

type CreateQuestResponse struct {
	ID string `json:"id"`
}

type GetQuestRequest struct {
	ID                       string `json:"id"`
	IncludeUnclaimableReason bool   `json:"include_unclaimable_reason"`
	EditMode                 bool   `json:"edit_mode"`
}

type GetQuestResponse Quest

type GetListQuestRequest struct {
	Q               string `json:"q"`
	CommunityHandle string `json:"community_handle"`
	CategoryID      string `json:"category"`
	Offset          int    `json:"offset"`
	Limit           int    `json:"limit"`

	IncludeUnclaimableReason bool `json:"include_unclaimable_reason"`
}

type GetListQuestResponse struct {
	Quests []Quest `json:"quests"`
	Hidden int     `json:"hidden"`
}

type GetQuestTemplatesRequest struct {
	Q      string `json:"q"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

type GetQuestTemplatestResponse struct {
	Templates []Quest `json:"templates"`
}

type ParseQuestTemplatesRequest struct {
	TemplateID      string `json:"template_id"`
	CommunityHandle string `json:"community_handle"`
}

type ParseQuestTemplatestResponse struct {
	Quest Quest `json:"quest"`
}

type UpdateQuestRequest struct {
	ID             string         `json:"id"`
	Status         string         `json:"status"`
	Type           string         `json:"type"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	CategoryID     string         `json:"category_id"`
	Recurrence     string         `json:"recurrence"`
	ValidationData map[string]any `json:"validation_data"`
	Points         uint64         `json:"points"`
	Rewards        []Reward       `json:"rewards"`
	ConditionOp    string         `json:"condition_op"`
	Conditions     []Condition    `json:"conditions"`
	IsHighlight    bool           `json:"is_highlight"`
}

type UpdateQuestResponse struct {
	Quest Quest `json:"quest"`
}

type DeleteQuestRequest struct {
	ID string `json:"id"`
}

type DeleteQuestResponse struct {
}
