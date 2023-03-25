package entity

type CollaboratorRole string

const (
	CollaboratorRoleReviewer CollaboratorRole = "Reviewer"
	CollaboratorRoleOwner    CollaboratorRole = "Owner"
	CollaboratorRoleEditor   CollaboratorRole = "Editor"
)

type Collaborator struct {
	Base
	UserID    string  `gorm:"not null"`
	ProjectID string  `gorm:"not null"`
	Project   Project `gorm:"foreignKey:ProjectID"`
	User      User    `gorm:"foreignKey:UserID"`
	Role      CollaboratorRole
	CreatedBy string `gorm:"not null"`
}

// func (e *Collaborator) Validate(db *gorm.DB) {
// 	var count int64
// 	db.Model(&User{}).Where("id = ?", e.UserID).Count(&count)
// 	if count == 0 {
// 		db.AddError(errors.New("user not found"))
// 	}

// 	db.Model(&Project{}).Where("id = ?", e.ProjectID).Count(&count)
// 	if count == 0 {
// 		db.AddError(errors.New("project not found"))
// 	}
// }
