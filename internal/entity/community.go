package entity

import (
	"database/sql"

	"github.com/questx-lab/backend/pkg/enum"
)

type ReferralStatusType string

var (
	ReferralUnclaimable = enum.New(ReferralStatusType("unclaimable"))
	ReferralPending     = enum.New(ReferralStatusType("pending"))
	ReferralClaimable   = enum.New(ReferralStatusType("claimable"))
	ReferralClaimed     = enum.New(ReferralStatusType("claimed"))
)

type CommunityStatus string

var (
	CommunityPending = enum.New(CommunityStatus("pending"))
	CommunityActive  = enum.New(CommunityStatus("active"))
)

type Community struct {
	Base
	CreatedBy      string
	CreatedByUser  User `gorm:"foreignKey:CreatedBy"`
	ReferredBy     sql.NullString
	ReferredByUser User `gorm:"foreignKey:ReferredBy"`
	ReferralStatus ReferralStatusType
	Handle         string `gorm:"unique"`
	DisplayName    string
	Followers      int
	TrendingScore  int
	LogoPicture    string
	Introduction   []byte `gorm:"type:longtext"`
	Twitter        string
	Discord        string
	WebsiteURL     string
	Status         CommunityStatus `gorm:"default:active"`
}
