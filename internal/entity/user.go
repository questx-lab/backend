package entity

import (
	"database/sql"

	"github.com/questx-lab/backend/pkg/enum"
)

type GlobalRole string

var (
	RoleSuperAdmin = enum.New(GlobalRole("super_admin"))
	RoleAdmin      = enum.New(GlobalRole("admin"))
	RoleUser       = enum.New(GlobalRole("user"))
)

var GlobalAdminRoles = []GlobalRole{RoleSuperAdmin, RoleAdmin}

type User struct {
	Base
	WalletAddress  sql.NullString `redis:"wallet_address" gorm:"unique"`
	Name           string         `redis:"name" gorm:"unique"`
	Role           GlobalRole     `redis:"role"`
	ProfilePicture string         `redis:"profile_picture"`
	ReferralCode   string         `redis:"referral_code"`
	IsNewUser      bool           `redis:"is_new_user"`
}
