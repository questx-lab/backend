package repository

import (
	"context"
	"log"
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/pkg/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_projectRepository_DeleteByID(t *testing.T) {
	ctx := context.Background()
	db := testutil.GetDatabaseTest()

	r := &projectRepository{
		db: db,
	}
	data := &entity.Project{
		Base: entity.Base{
			ID: "valid-project-id",
		},
	}
	db.Model(&entity.Project{}).Create(data)
	err := r.DeleteByID(ctx, data.ID)
	assert.NoError(t, err)
	var p entity.Project
	// tx := db.Model(&entity.Project{}).Where("id = ?", data.ID).First(&p)
	tx := db.Model(&entity.Project{}).Unscoped().Where("deleted_at IS NOT NULL").Take(&p, "id", data.ID)
	log.Println(tx.Statement.SQL.String())
	err = tx.Error
	assert.NoError(t, err)
	assert.Equal(t, data.ID, p.ID)
	assert.NotEmpty(t, p.DeletedAt)
}
