package entity

type User struct {
	ID      string `gorm:"primaryKey"`
	Address string
	Name    string
}
