package config

import (
	"fmt"
	"time"

	"github.com/questx-lab/backend/utils/token"
)

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
	DB         *Database
	Port       string
	Server     ServerConfigs
	Auth       AuthConfigs
	TknConfigs token.Configs
}

type ServerConfigs struct {
	Host string
	Port string
	Cert string
	Key  string
}

type AuthConfigs struct {
	TokenSecret     string
	TokenExpiration time.Duration
	AccessTokenName string
	SessionSecret   string

	Google  OAuth2Config
	Twitter OAuth2Config
}

type OAuth2Config struct {
	Name         string
	Issuer       string
	ClientID     string
	ClientSecret string
	IDField      string
}
