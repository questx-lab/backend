package cqlutil

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTableNames(t *testing.T) {
	type test struct {
		Name                  string
		LongNameWithCamelCase string
		Somethingwrong        string
	}
	got := GetTableNames(&test{})

	want := []string{"name", "long_name_with_camel_case", "somethingwrong"}

	sort.Strings(want)
	require.Equal(t, want, got)

}
