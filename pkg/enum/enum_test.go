package enum

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("create a enum of string", func(t *testing.T) {
		type EnumString string

		bar := New(EnumString("bar"))
		require.Equal(t, bar, EnumString("bar"))

		v, err := ToEnum[EnumString]("bar")
		require.NoError(t, err)
		require.Equal(t, v, bar)

		require.Equal(t, string(bar), "bar")
	})
}
