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
	WalletAddress  sql.NullString `gorm:"unique"`
	Name           string         `gorm:"unique"`
	Role           GlobalRole
	ProfilePicture string
	ReferralCode   string
	IsNewUser      bool
}
