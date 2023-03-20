package config

type Configs struct {
	DBConnection string
	Port         string

	JwtSecretKey string
	JwtExpiredAt int64
}
