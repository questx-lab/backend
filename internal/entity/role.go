package entity

import "database/sql"

type PermissionFlag uint64

const (
	DELETE_COMMUNITY PermissionFlag = 1 << iota
	EDIT_COMMUNITY
	MANAGE_QUEST
	REVIEW_CLAIMED_QUEST
	MANAGE_CHANNEL
	MANAGE_ROLE
	KICK_MEMBER
	BAN_MEMBER
	TIMEOUT_MEMBER
	ASSIGN_ROLE
	MANAGE_LOTTERY
)

type Role struct {
	Base
	CommunityID sql.NullString
	Community   Community `gorm:"foreignKey:CommunityID"`
	Name        string
	Permissions uint64
	Priority    int
	Color       string
}

// role-based access control
var RBAC = map[string]PermissionFlag{
	"/deleteCommunity":        DELETE_COMMUNITY,
	"/updateCommunity":        EDIT_COMMUNITY,
	"/updateCommunityDiscord": EDIT_COMMUNITY,
	"/uploadCommunityLogo":    EDIT_COMMUNITY,
	"/updateBadge":            EDIT_COMMUNITY,
	"/getMyReferrals":         EDIT_COMMUNITY,
	"/createQuest":            MANAGE_QUEST,
	"/updateQuest":            MANAGE_QUEST,
	"/updateQuestCategory":    MANAGE_QUEST,
	"/updateQuestPosition":    MANAGE_QUEST,
	"/deleteQuest":            MANAGE_QUEST,
	"/createCategory":         MANAGE_QUEST,
	"/updateCategory":         MANAGE_QUEST,
	"/deleteCategory":         MANAGE_QUEST,
	"/getClaimedQuest":        REVIEW_CLAIMED_QUEST,
	"/getClaimedQuests":       REVIEW_CLAIMED_QUEST,
	"/review":                 REVIEW_CLAIMED_QUEST,
	"/reviewAll":              REVIEW_CLAIMED_QUEST,
	"/givePoint":              REVIEW_CLAIMED_QUEST,
	"/createChannel":          MANAGE_CHANNEL,
	"/createLotteryEvent":     MANAGE_LOTTERY,
	"/deleteMessage":          MANAGE_CHANNEL,
	"/createRole":             MANAGE_ROLE,
	"/updateRole":             MANAGE_ROLE,
	"/deleteRole":             MANAGE_ROLE,
	"/assignCommunityRole":    ASSIGN_ROLE,
}

const (
	UserBaseRole  string = "user"
	OwnerBaseRole string = "owner"
)
