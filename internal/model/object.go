package model

type AccessToken struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type RefreshToken struct {
	Family  string
	Counter uint64
}

type Category struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	CreatedBy   string `json:"created_by,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

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

type Collaborator struct {
	ProjectID string  `json:"project_id,omitempty"`
	Project   Project `json:"project,omitempty"`
	UserID    string  `json:"user_id,omitempty"`
	User      User    `json:"user,omitempty"`
	Role      string  `json:"name,omitempty"`
	CreatedBy string  `json:"created_by,omitempty"`
}

type Project struct {
	ID        string `json:"id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`

	CreatedBy    string `json:"created_by,omitempty"`
	Introduction string `json:"introduction,omitempty"`
	Name         string `json:"name,omitempty"`
	Twitter      string `json:"twitter,omitempty"`
	Discord      string `json:"discord,omitempty"`
}

type Reward struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

type Condition struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

type Quest struct {
	ID                string         `json:"id,omitempty"`
	ProjectID         string         `json:"project_id,omitempty"`
	Type              string         `json:"type,omitempty"`
	Status            string         `json:"status,omitempty"`
	Title             string         `json:"title,omitempty"`
	Description       string         `json:"description,omitempty"`
	Categories        []string       `json:"categories,omitempty"`
	Recurrence        string         `json:"recurrence,omitempty"`
	ValidationData    map[string]any `json:"validation_data,omitempty"`
	Rewards           []Reward       `json:"rewards,omitempty"`
	ConditionOp       string         `json:"condition_op,omitempty"`
	Conditions        []Condition    `json:"conditions,omitempty"`
	CreatedAt         string         `json:"created_at,omitempty"`
	UpdatedAt         string         `json:"updated_at,omitempty"`
	UnclaimableReason string         `json:"unclaimable_reason,omitempty"`
}

type UserAggregate struct {
	UserID      string `json:"user_id"`
	TotalTask   uint64 `json:"total_task"`
	TotalPoint  uint64 `json:"total_point"`
	PrevRank    uint64 `json:"prev_rank"`
	CurrentRank uint64 `json:"current_rank"`
}

type User struct {
	ID       string            `json:"id,omitempty"`
	Address  string            `json:"address,omitempty"`
	Name     string            `json:"name,omitempty"`
	Role     string            `json:"role,omitempty"`
	Services map[string]string `json:"services,omitempty"`
}

type Participant struct {
	UserID      string `json:"user_id,omitempty"`
	Points      uint64 `json:"points,omitempty"`
	InviteCode  string `json:"invite_code,omitempty"`
	InvitedBy   string `json:"invited_by,omitempty"`
	InviteCount uint64 `json:"invite_count,omitempty"`
}

type Badge struct {
	UserID      string `json:"user_id,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Level       int    `json:"level,omitempty"`
	WasNotified bool   `json:"was_notified,omitempty"`
}
