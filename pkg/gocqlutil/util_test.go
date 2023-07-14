package gocqlutil

import (
	"sort"
	"testing"

	"github.com/questx-lab/backend/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestGetTableNames(t *testing.T) {
	got := GetTableNames(&entity.ChatChannel{})

	want := []string{"id", "name", "created_at", "created_by", "deleted_at", "updated_at", "community_id"}

	sort.Strings(want)
	require.Equal(t, want, got)

}
