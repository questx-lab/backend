package entity

type Project struct {
	Base
	CreatedBy string
	Name      string
	Twitter   string
	Discord   string
	Telegram  string
}

func (e *Project) Table() string {
	return "projects"
}
