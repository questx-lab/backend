package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/questx-lab/backend/internal/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
	suite.Suite
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

func (suite *UserTestSuite) TestReadWriteUser() {
	t := suite.T()
	// In-memory test
	testReadWriteUser(t, testutil.GetEmptyTestDB(t))

	// Real DB test
	if testutil.EnableIntegrationTest() {
		db := testutil.GetEmptyIntegrationDb(t)
		testReadWriteUser(t, db)
		db.Close()
	}
}

func testReadWriteUser(t *testing.T, db *sql.DB) {
	userRepo := NewUserRepository(db)

	err := userRepo.Create(context.Background(), &entity.User{
		ID:   "id1",
		Name: "user1",
	})
	require.Nil(t, err)

	user, err := userRepo.RetrieveByID(context.Background(), "id1")
	require.Nil(t, err)
	require.Equal(t, "user1", user.Name)
}
