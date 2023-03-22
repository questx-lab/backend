package migration

import (
	"github.com/questx-lab/backend/internal/entity"
	"gorm.io/gorm"
)

func DoMigration(db *gorm.DB) {
	db.AutoMigrate(&entity.User{}, &entity.OAuth2{})
}
