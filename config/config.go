package config

import "time"

type Configs struct {
	Database DatabaseConfig
	Server   ServerConfigs
	Auth     AuthConfigs
}

type DatabaseConfig struct {
	DSN string
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
