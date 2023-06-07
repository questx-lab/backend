package entity

type Badge struct {
	Base
	Name        string `gorm:"index:idx_badges_name_level,unique"`
	Level       int    `gorm:"index:idx_badges_name_level,unique"`
	Description string
	Value       int
	IconURL     string
}
