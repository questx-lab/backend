package entity

import "github.com/questx-lab/backend/pkg/enum"

type GlobalRole string

var (
	RoleSuperAdmin = enum.New(GlobalRole("super_admin"))
	RoleAdmin      = enum.New(GlobalRole("admin"))
	RoleUser       = enum.New(GlobalRole("user"))
)

var GlobalAdminRole = []GlobalRole{RoleSuperAdmin, RoleAdmin}

type User struct {
	Base
	Address string `gorm:"unique"`
	Name    string `gorm:"unique"`
	Role    GlobalRole
}
