package config

import (
	"fmt"
	"time"

	"github.com/questx-lab/backend/pkg/storage"
)

type Configs struct {
	Env string

	Database DatabaseConfigs
	Server   ServerConfigs
	Auth     AuthConfigs
	Session  SessionConfigs
	Storage  storage.S3Configs
	File     FileConfigs
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
	CallbackURL  string
	TokenSecret  string
	AccessToken  TokenConfigs
	RefreshToken TokenConfigs

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
	Name       string
	Expiration time.Duration
}

type FileConfigs struct {
	MaxSize int
}
