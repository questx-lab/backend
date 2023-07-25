package reflectutil

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetColumnNames(t *testing.T) {
	type test struct {
		Name                  string
		LongNameWithCamelCase string
		Somethingwrong        string
		UserIDs               string
	}
	got := GetColumnNames(&test{})

	want := []string{"name", "long_name_with_camel_case", "somethingwrong", "user_ids"}

	sort.Strings(want)
	require.Equal(t, want, got)

}
