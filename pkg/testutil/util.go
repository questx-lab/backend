package testutil

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func GetDatabaseTest() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}
