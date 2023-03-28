package config

import (
	"fmt"
	"time"
)

type Configs struct {
	Env      string
	Database DatabaseConfigs
	Server   ServerConfigs
	Auth     AuthConfigs
	Token    TokenConfigs
	Session  SessionConfigs
}

type DatabaseConfigs struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

func (d *DatabaseConfigs) ConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
	)
}

type ServerConfigs struct {
	Host string
	Port string
	Cert string
	Key  string
}

type SessionConfigs struct {
	Secret string
	Name   string
}

type AuthConfigs struct {
	AccessTokenName string
	CallbackURL     string

	Google OAuth2Config
}

type OAuth2Config struct {
	Name         string
	Issuer       string
	ClientID     string
	ClientSecret string
	IDField      string
}

type TokenConfigs struct {
	Expiration time.Duration
	Secret     string
}
