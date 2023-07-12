package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"hash"
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
		b[i] = alphabet[RandIntn(len(alphabet))]
	}
	return string(b)
}

func SHA256(b []byte) string {
	hashed := sha256.Sum256(b)
	return base64.StdEncoding.EncodeToString(hashed[:])
}

func HMAC(hashFunc func() hash.Hash, data []byte, secret []byte) string {
	h := hmac.New(hashFunc, secret)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// RandIntn returns a uniform random value in [0, n). It panics if got a
// non-positive parameter.
func RandIntn(n int) int {
	r, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(err)
	}

	return int(r.Int64())
}

// RandRange returns a uniform random value in [a, b). It panics if got a
// non-positive parameter or a>=b.
func RandRange(a, b int) int {
	return RandIntn(b-a) + a
}
