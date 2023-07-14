package entity

import "database/sql"

type PermissionFlag int

const (
	MANAGE_COMMUNITY     PermissionFlag = 0
	MANAGE_QUEST         PermissionFlag = 1
	REVIEW_CLAIMED_QUEST PermissionFlag = 2
)

type Role struct {
	Base
	CommunityID sql.NullString
	Community   Community `gorm:"foreignKey:CommunityID"`
	Name        string
	Permissions int64
}
