package entity

type User struct {
	Base
	Address string
	Name    string `gorm:"unique"`
	Role    string `gorm:"default:USER"`
}

const (
	SuperAdminRole = "SUPER_ADMIN"
	AdminRole      = "ADMIN"
	UserRole       = "USER"
)
