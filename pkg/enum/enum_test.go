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

		v, err := ToEnum[EnumString]("Bar")
		require.NoError(t, err)
		require.Equal(t, v, bar)

		_, err = ToEnum[EnumString]("bar")
		require.Error(t, err)

		require.Equal(t, ToString(bar), "Bar")
		require.Equal(t, ToString(EnumString("foo")), "")
	})

	t.Run("create a enum of int", func(t *testing.T) {
		type EnumInt int

		bar := New(EnumInt(100), "Bar")
		require.Equal(t, bar, EnumInt(100))

		v, err := ToEnum[EnumInt]("Bar")
		require.NoError(t, err)
		require.Equal(t, v, bar)

		_, err = ToEnum[EnumInt]("bar")
		require.Error(t, err)

		require.Equal(t, ToString(bar), "Bar")
		require.Equal(t, ToString(EnumInt(200)), "")
	})
}
