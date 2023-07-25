package model

import "github.com/questx-lab/backend/internal/entity"

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
	ID             string `json:"id"`
	Quest          Quest  `json:"quest"`
	User           User   `json:"user"`
	Status         string `json:"status"`
	SubmissionData string `json:"submission_data"`
	ReviewerID     string `json:"reviewer_id"`
	ReviewedAt     string `json:"reviewed_at"`
	Comment        string `json:"comment"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
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
	WebsiteURL     string `json:"website_url"`
	Status         string `json:"status"`
	OwnerEmail     string `json:"owner_email"`

	Channels []ChatChannel `json:"channels,omitempty"`
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
	Community         Community      `json:"community"`
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
	Position          int            `json:"position"`
}

type User struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	WalletAddress      string            `json:"wallet_address"`
	Role               string            `json:"role"`
	Services           map[string]string `json:"services"`
	ReferralCode       string            `json:"referral_code"`
	IsNewUser          bool              `json:"is_new_user"`
	AvatarURL          string            `json:"avatar_url"`
	TotalCommunities   int               `json:"total_communities"`
	TotalClaimedQuests int               `json:"total_claimed_quests"`
}

type Role struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Permision uint64 `json:"permision"`
}

type Follower struct {
	User        User      `json:"user"`
	Community   Community `json:"community"`
	Roles       []Role    `json:"role"`
	Points      uint64    `json:"points"`
	Quests      uint64    `json:"quests"`
	Streaks     uint64    `json:"streaks"`
	InviteCode  string    `json:"invite_code"`
	InvitedBy   string    `json:"invited_by"`
	InviteCount uint64    `json:"invite_count"`
}

type Badge struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
}

type BadgeDetail struct {
	User        User      `json:"user"`
	Community   Community `json:"community"`
	Badge       Badge     `json:"badge"`
	WasNotified bool      `json:"was_notified"`
	CreatedAt   string    `json:"created_at"`
}

type BlockchainConnection struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type Blockchain struct {
	Name                  string                 `json:"name"`
	ID                    int64                  `json:"id"`
	UseExternalRPC        bool                   `json:"use_external_rpc"`
	UseEip1559            bool                   `json:"use_eip_1559"`
	BlockTime             int                    `json:"block_time"`
	AdjustTime            int                    `json:"adjust_time"`
	ThresholdUpdateBlock  int                    `json:"threshold_update_block"`
	BlockchainConnections []BlockchainConnection `json:"blockchain_connections"`
}

type BlockchainToken struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Chain   string `json:"chain"`
	Symbol  string `json:"symbol"`
	Address string `json:"address"`
}

type BlockchainTransaction struct {
	TxHash    string `json:"tx_hash"`
	Chain     string `json:"chain"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`

	PayRewards []PayReward `json:"pay_rewards"`
}

type PayReward struct {
	ID                      string                `json:"id"`
	Token                   BlockchainToken       `json:"token"`
	ClaimedQuestID          string                `json:"claimed_quest_id"`
	ReferralCommunityHandle string                `json:"referral_community_handle"`
	FromCommunityHandle     string                `json:"from_community_handle"`
	ToUser                  User                  `json:"to_user"`
	ToAddress               string                `json:"to_address"`
	Amount                  float64               `json:"amount"`
	CreatedAt               string                `json:"created_at"`
	UpdatedAt               string                `json:"updated_at"`
	Transaction             BlockchainTransaction `json:"transaction"`
}

type UserStatistic struct {
	User         User `json:"user"`
	Value        int  `json:"value"`
	CurrentRank  int  `json:"current_rank"`
	PreviousRank int  `json:"previous_rank"`
}

type Referral struct {
	ReferredBy  User        `json:"referred_by"`
	Communities []Community `json:"communities"`
}

type GameMap struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ConfigURL string `json:"config_url"`
}

type GameRoom struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Map  GameMap `json:"map"`
}

type GameCharacter struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Level             int     `json:"level"`
	ConfigURL         string  `json:"config_url"`
	ImageURL          string  `json:"image_url"`
	ThumbnailURL      string  `json:"thumbnail_url"`
	SpriteWidthRatio  float64 `json:"sprite_width_ratio"`
	SpriteHeightRatio float64 `json:"sprite_height_ratio"`
	Points            int     `json:"points"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type GameCommunityCharacter struct {
	CommunityID   string        `json:"community_id"`
	Points        int           `json:"points"`
	GameCharacter GameCharacter `json:"game_character"`
	CreatedAt     string        `json:"created_at"`
	UpdatedAt     string        `json:"updated_at"`
}

type GameUserCharacter struct {
	UserID        string        `json:"user_id"`
	CommunityID   string        `json:"community_id"`
	IsEquipped    bool          `json:"is_equipped"`
	GameCharacter GameCharacter `json:"game_character"`
	CreatedAt     string        `json:"created_at"`
	UpdatedAt     string        `json:"updated_at"`
}

type DiscordRole struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

type ChatMessage struct {
	ID          int64               `json:"id"`
	ChannelID   int64               `json:"channel_id"`
	Author      User                `json:"author"`
	Content     string              `json:"content"`
	ReplyTo     int64               `json:"reply_to,omitempty"`
	Attachments []entity.Attachment `json:"attachments,omitempty"`
	Reactions   []ChatReactionState `json:"reactions,omitempty"`
}

type ChatReactionState struct {
	Emoji entity.Emoji `json:"emoji"`
	Count int          `json:"count"`
	Me    bool         `json:"me"`
}

type ChatChannel struct {
	ID              int64  `json:"id"`
	UpdatedAt       string `json:"updated_at"`
	CommunityHandle string `json:"community_handle"`
	Name            string `json:"name"`
	LastMessageID   int64  `json:"last_message_id"`
}

type ChatMember struct {
	UserID            string      `json:"user_id"`
	Channel           ChatChannel `json:"channel"`
	LastReadMessageID int64       `json:"last_read_message_id"`
}

type LotteryPrize struct {
	ID               string   `json:"id"`
	EventID          string   `json:"event_id"`
	Points           int      `json:"points"`
	Rewards          []Reward `json:"rewards"`
	AvailableRewards int      `json:"available_rewards"`
}

type LotteryEvent struct {
	ID             string         `json:"id"`
	Community      Community      `json:"community"`
	StartTime      string         `json:"start_time"`
	EndTime        string         `json:"end_time"`
	MaxTickets     int            `json:"max_tickets"`
	UsedTickets    int            `json:"used_tickets"`
	PointPerTicket int            `json:"point_per_ticket"`
	Prizes         []LotteryPrize `json:"prizes"`
}

type LotteryWinner struct {
	ID        string       `json:"id"`
	CreatedAt string       `json:"created_at"`
	Prize     LotteryPrize `json:"prize"`
	User      User         `json:"user"`
}
