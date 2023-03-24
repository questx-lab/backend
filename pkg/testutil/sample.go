package testutil

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/repository"
	"gorm.io/gorm"
)

// SamepleProject creates a new project in database with many fields are
// randomized. The sample project can be overwritten by non-zero fields of init.
//
// This function returns the sample project.
func SampleProject(db *gorm.DB, init *entity.Project) (entity.Project, error) {
	projectRepo := repository.NewProjectRepository(db)

	sample := &entity.Project{
		Base:      entity.Base{ID: uuid.NewString()},
		Name:      uuid.NewString(),
		CreatedBy: uuid.NewString(),
		Twitter:   "https://twitter.com/hashtag/Breaking2",
		Discord:   "https://discord.com/hashtag/Breaking2",
		Telegram:  "https://telegram.com",
	}

	if init != nil {
		overwriteFields(sample, *init)
	}

	if err := projectRepo.Create(context.Background(), sample); err != nil {
		return *sample, err
	}
	return *sample, nil
}

func overwriteFields[T any](origin *T, overwrite T) {
	originValue := reflect.ValueOf(origin).Elem()
	overwriteValue := reflect.ValueOf(overwrite)

	for i := 0; i < overwriteValue.NumField(); i++ {
		overwriteField := overwriteValue.Field(i)
		if overwriteField.Interface() != reflect.Zero(overwriteField.Type()).Interface() {
			originValue.Field(i).Set(overwriteField)
		}
	}
}
