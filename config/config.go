package config

import "fmt"

type Database struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

func (d *Database) ConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
	)
}

type Configs struct {
	DB   *Database
	Port string

	JwtSecretKey string
	JwtExpiredAt int64
}
