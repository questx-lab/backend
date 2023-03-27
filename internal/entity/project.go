package entity

type Project struct {
	Base
	CreatedBy string `gorm:"not null"`
	Name      string
	Twitter   string
	Discord   string
	Telegram  string
}
