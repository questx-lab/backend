package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"math/big"
)

func GenerateRandomString() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateRandomAlphabet(n uint) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[CryptoRandIntn(len(alphabet))]
	}
	return string(b)
}

func Hash(b []byte) string {
	hashed := sha256.Sum224(b)
	return base64.StdEncoding.EncodeToString(hashed[:])
}

// CryptoRandIntn returns a uniform random value in [0, n). It panics if got a negative parameter.
func CryptoRandIntn(n int) int {
	r, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(err)
	}

	return int(r.Int64())
}
