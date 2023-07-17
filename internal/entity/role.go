package entity

import "database/sql"

type PermissionFlag int

const (
	OWNER_COMMUNITY PermissionFlag = iota
	EDITOR_COMMUNITY
	MANAGE_QUEST
	REVIEW_CLAIMED_QUEST
	MANAGE_CHANNEL
	MANAGE_ROLE
	KICK_MEMBER
	BAN_MEMBER
	TIMEOUT_MEMBER
)

type Role struct {
	Base
	CommunityID sql.NullString
	Community   Community `gorm:"foreignKey:CommunityID"`
	Name        string
	Permissions int64
}

// role-based access control
var RBAC = map[string][]PermissionFlag{
	"/deleteCommunity":        {OWNER_COMMUNITY},
	"/claimReferral":          {OWNER_COMMUNITY},
	"/updateCommunity":        {EDITOR_COMMUNITY},
	"/updateCommunityDiscord": {EDITOR_COMMUNITY},
	"/uploadCommunityLogo":    {EDITOR_COMMUNITY},
	"/updateBadge":            {EDITOR_COMMUNITY},
	"/getMyReferrals":         {EDITOR_COMMUNITY},
	"/createQuest":            {MANAGE_QUEST},
	"/updateQuest":            {MANAGE_QUEST},
	"/updateQuestCategory":    {MANAGE_QUEST},
	"/updateQuestPosition":    {MANAGE_QUEST},
	"/deleteQuest":            {MANAGE_QUEST},
	"/createCategory":         {MANAGE_QUEST},
	"/updateCategory":         {MANAGE_QUEST},
	"/deleteCategory":         {MANAGE_QUEST},
	"/getClaimedQuest":        {REVIEW_CLAIMED_QUEST},
	"/getClaimedQuests":       {REVIEW_CLAIMED_QUEST},
	"/review":                 {REVIEW_CLAIMED_QUEST},
	"/reviewAll":              {REVIEW_CLAIMED_QUEST},
	"/givePoint":              {REVIEW_CLAIMED_QUEST},
}

type BaseRole string

const (
	UserBaseRole     BaseRole = "user"
	OwnerBaseRole    BaseRole = "owner"
	EditorBaseRole   BaseRole = "editor"
	ReviewerBaseRole BaseRole = "reviewer"
)
