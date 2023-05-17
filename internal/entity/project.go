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

type Project struct {
	Base
	CreatedBy      string
	CreatedByUser  User `gorm:"foreignKey:CreatedBy"`
	ReferredBy     sql.NullString
	ReferredByUser User `gorm:"foreignKey:ReferredBy"`
	ReferralStatus ReferralStatusType
	Name           string `gorm:"unique"`
	Followers      int
	LogoPictures   Map    // Contains images in different sizes.
	Introduction   []byte `gorm:"type:longtext"`
	Twitter        string
	Discord        string

	WebsiteURL         string
	DevelopmentStage   string
	TeamSize           int
	SharedContentTypes Array[string]
}
