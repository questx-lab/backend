package token

import (
	"testing"

	"github.com/questx-lab/backend/config"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	configs := &config.Configs{
		JwtSecretKey: "abcxyz",
		JwtExpiredAt: 300000000,
	}
	id := "0111223xz"
	token, err := Generate(id, configs)
	assert.NoError(t, err)

	verifyID, err := Verify(token, configs)
	assert.NoError(t, err)

	assert.Equal(t, id, verifyID)
}
