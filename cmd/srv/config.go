package main

import "time"

type Configs struct {
	DBConnection  string
	SessionSecret string
	Server        ServerConfigs
	Auth          AuthConfigs
}

type ServerConfigs struct {
	Address string
	Port    string
	Cert    string
	Key     string
}

type AuthConfigs struct {
	TokenSecret     string
	TokenExpiration time.Duration
	Google          OAuth2Config
	Tweeter         OAuth2Config
}

type OAuth2Config struct {
	Name         string
	Issuer       string
	ClientID     string
	ClientSecret string
}
