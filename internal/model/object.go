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
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ClaimedQuest struct {
	ID         string `json:"id"`
	Quest      Quest  `json:"quest"`
	User       User   `json:"user"`
	Status     string `json:"status"`
	Input      string `json:"input"`
	ReviewerID string `json:"reviewer_id"`
	ReviewedAt string `json:"reviewed_at"`
	Comment    string `json:"comment"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type Collaborator struct {
	Community Community `json:"community"`
	User      User      `json:"user"`
	Role      string    `json:"name"`
	CreatedBy string    `json:"created_by"`
}

type Community struct {
	Handle    string `json:"handle"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`

	ReferredBy     string `json:"referred_by"`
	ReferralStatus string `json:"referral_status"`
	CreatedBy      string `json:"created_by"`
	Introduction   string `json:"introduction"`
	DisplayName    string `json:"display_name"`
	Twitter        string `json:"twitter"`
	Discord        string `json:"discord"`
	Followers      int    `json:"followers"`
	NumberOfQuests int    `json:"number_of_quests"`
	TrendingScore  int    `json:"trending_score"`
	LogoURL        string `json:"logo_url"`

	WebsiteURL         string   `json:"website_url"`
	DevelopmentStage   string   `json:"development_stage"`
	TeamSize           int      `json:"team_size"`
	SharedContentTypes []string `json:"shared_content_types"`
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
	ID                string         `json:"id"`
	Community         Community      `json:"community_handle"`
	Type              string         `json:"type"`
	Status            string         `json:"status"`
	Title             string         `json:"title"`
	Description       string         `json:"description"`
	Category          Category       `json:"category"`
	Recurrence        string         `json:"recurrence"`
	ValidationData    map[string]any `json:"validation_data"`
	Points            uint64         `json:"points"`
	Rewards           []Reward       `json:"rewards"`
	ConditionOp       string         `json:"condition_op"`
	Conditions        []Condition    `json:"conditions"`
	CreatedAt         string         `json:"created_at"`
	UpdatedAt         string         `json:"updated_at"`
	UnclaimableReason string         `json:"unclaimable_reason"`
	IsHighlight       bool           `json:"is_highlight"`
}

type User struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	WalletAddress string            `json:"wallet_address"`
	Role          string            `json:"role"`
	Services      map[string]string `json:"services"`
	ReferralCode  string            `json:"referral_code"`
	IsNewUser     bool              `json:"is_new_user"`
	AvatarURL     string            `json:"avatar_url"`
}

type Follower struct {
	User        User      `json:"user"`
	Community   Community `json:"community"`
	Points      uint64    `json:"points"`
	Quests      uint64    `json:"quests"`
	Streaks     uint64    `json:"streaks"`
	InviteCode  string    `json:"invite_code"`
	InvitedBy   string    `json:"invited_by"`
	InviteCount uint64    `json:"invite_count"`
}

type Badge struct {
	User        User      `json:"user"`
	Community   Community `json:"community"`
	Name        string    `json:"name"`
	Level       int       `json:"level"`
	WasNotified bool      `json:"was_notified"`
}

type Transaction struct {
	ID             string  `json:"id"`
	User           User    `json:"user"`
	CreatedAt      string  `json:"created_at"`
	ClaimedQuestID string  `json:"claimed_quest_id"`
	Note           string  `json:"note"`
	Status         string  `json:"status"`
	Address        string  `json:"address"`
	Token          string  `json:"token"`
	Amount         float64 `json:"amount"`
}

type UserStatistic struct {
	User         User `json:"user"`
	Value        int  `json:"value"`
	CurrentRank  int  `json:"current_rank"`
	PreviousRank int  `json:"previous_rank"`
}
