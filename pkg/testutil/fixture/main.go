package main

import (
	"context"
	"fmt"

	"github.com/questx-lab/backend/internal/repository"
	"github.com/questx-lab/backend/pkg/testutil"
	"gorm.io/gorm"
)

func read(db *gorm.DB) {
	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.GetByID(context.Background(), "user3")
	if err != nil {
		panic(err)
	}

	fmt.Println("user = ", user)
}

func main() {
	testutil.CreateFixture()

	db := testutil.DefaultTestDb()
	read(db)
}
