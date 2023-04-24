package entity

type User struct {
	Base
	Address string
	Name    string `gorm:"unique"`
	Role    string
}

const (
	SuperAdminRole = "SUPER_ADMIN"
	AdminRole      = "ADMIN"
	UserRole       = "USER"
)
