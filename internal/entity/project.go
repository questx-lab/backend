package entity

type Project struct {
	Base
	CreatedBy string `gorm:"->"`
	Name      string
	Twitter   string
	Discord   string
	Telegram  string
}
