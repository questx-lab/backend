package entity

import "github.com/questx-lab/backend/pkg/enum"

type GlobalRole string

var (
	RoleSuperAdmin = enum.New(GlobalRole("super_admin"))
	RoleAdmin      = enum.New(GlobalRole("admin"))
	RoleUser       = enum.New(GlobalRole("user"))
)

// TODO: For testing purpose, we will delete this role group when we can assign
// global role to user.
var AnyGlobalRole = []GlobalRole{RoleSuperAdmin, RoleAdmin, RoleUser}

type User struct {
	Base
	Address string
	Name    string     `gorm:"unique"`
	Role    GlobalRole `gorm:"default:user"`
}
