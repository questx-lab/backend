package entity

import (
	"database/sql"
	"time"

	"github.com/questx-lab/backend/pkg/enum"
)

type ReferralStatusType string

var (
	ReferralUnclaimable = enum.New(ReferralStatusType("unclaimable"))
	ReferralPending     = enum.New(ReferralStatusType("pending"))
	ReferralClaimable   = enum.New(ReferralStatusType("claimable"))
	ReferralClaimed     = enum.New(ReferralStatusType("claimed"))
	ReferralRejected    = enum.New(ReferralStatusType("rejected"))
)

type CommunityStatus string

var (
	CommunityPending  = enum.New(CommunityStatus("pending"))
	CommunityActive   = enum.New(CommunityStatus("active"))
	CommunityRejected = enum.New(CommunityStatus("rejected"))
)

type Community struct {
	Base
	CreatedBy         string
	CreatedByUser     User `gorm:"foreignKey:CreatedBy"`
	ReferredBy        sql.NullString
	ReferredByUser    User `gorm:"foreignKey:ReferredBy"`
	ReferralStatus    ReferralStatusType
	Handle            string `gorm:"unique"`
	DisplayName       string
	Followers         int
	TrendingScore     int
	LogoPicture       string
	Introduction      []byte `gorm:"type:longtext"`
	Twitter           string
	Discord           string
	DiscordInviteLink string
	WebsiteURL        string
	Status            CommunityStatus
	OwnerEmail        string
	WalletNonce       string
}

type CommunityRecord struct {
	CommunityID string    `gorm:"primaryKey"`
	Community   Community `gorm:"foreignKey:CommunityID"`

	Date time.Time `gorm:"primaryKey"`

	Followers int
}
