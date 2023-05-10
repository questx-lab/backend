package entity

import "github.com/questx-lab/backend/pkg/enum"

type Role string

var (
	Reviewer = enum.New(Role("reviewer"))
	Owner    = enum.New(Role("owner"))
	Editor   = enum.New(Role("editor"))
)

var ReviewGroup = []Role{Owner, Editor, Reviewer}
var AdminGroup = []Role{Owner, Editor}

type Collaborator struct {
	UserID        string  `gorm:"primaryKey"`
	User          User    `gorm:"foreignKey:UserID"`
	ProjectID     string  `gorm:"primaryKey"`
	Project       Project `gorm:"foreignKey:ProjectID"`
	Role          Role
	CreatedBy     string
	CreatedByUser User `gorm:"foreignKey:CreatedBy"`
}

func (c *Collaborator) TableName() string {
	return "collaborators"
}
