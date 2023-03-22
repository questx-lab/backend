package token

type Configs struct {
	JwtSecretKey string
	JwtExpiredAt int64
}

type Generator interface {
	Generate(id string) (string, error)
	Verify(token string) (string, error)
}
