package entity

import (
	"database/sql"

	"github.com/questx-lab/backend/pkg/enum"
)

type InvitedStatusType string

var (
	InvitedStatusUnclaimable = enum.New(InvitedStatusType("unclaimable"))
	InvitedStatusPending     = enum.New(InvitedStatusType("pending"))
	InvitedStatusClaimable   = enum.New(InvitedStatusType("claimable"))
	InvitedStatusClaimed     = enum.New(InvitedStatusType("claimed"))
)

type Community struct {
	Base
	CreatedBy     string
	CreatedByUser User `gorm:"foreignKey:CreatedBy"`
	InvitedBy     sql.NullString
	InvitedByUser User `gorm:"foreignKey:InvitedBy"`
	InvitedStatus InvitedStatusType
	Handle        string `gorm:"unique"`
	DisplayName   string
	Followers     int
	TrendingScore int
	LogoPicture   string
	Introduction  []byte `gorm:"type:longtext"`
	Twitter       string
	Discord       string
	WebsiteURL    string
}
