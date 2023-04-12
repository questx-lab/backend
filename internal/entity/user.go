package entity

type User struct {
	Base
	Address string
	Name    string `gorm:"unique"`
}
