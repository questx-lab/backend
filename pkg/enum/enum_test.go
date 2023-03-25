package enum

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("create a enum of string", func(t *testing.T) {
		type EnumString string

		bar := New(EnumString("bar"), "Bar")

		require.Equal(t, bar, EnumString("bar"))
		require.Equal(t, ToEnum[EnumString]("Bar"), bar)
		require.Equal(t, ToEnum[EnumString]("bar"), EnumString(""))

		require.Equal(t, ToString(bar), "Bar")
		require.Equal(t, ToString(EnumString("foo")), "")
	})

	t.Run("create a enum of int", func(t *testing.T) {
		type EnumInt int

		bar := New(EnumInt(100), "Bar")

		require.Equal(t, bar, EnumInt(100))
		require.Equal(t, ToEnum[EnumInt]("Bar"), bar)
		require.Equal(t, ToEnum[EnumInt]("bar"), EnumInt(0))

		require.Equal(t, ToString(bar), "Bar")
		require.Equal(t, ToString(EnumInt(200)), "")
	})
}
