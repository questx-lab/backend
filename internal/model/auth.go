package model

import "net/http"

// Access Token and Refresh Token
type AccessToken struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type RefreshToken struct {
	ID string `json:"id"`
}

// OAuth2 Login
type OAuth2LoginRequest struct {
	Type string `json:"type"`
}

type OAuth2LoginResponse struct {
	RedirectURL string `json:"-"`
	State       string `json:"-"`
}

func (r OAuth2LoginResponse) RedirectInfo() (int, string) {
	return http.StatusTemporaryRedirect, r.RedirectURL
}

func (r OAuth2LoginResponse) SessionInfo() map[string]any {
	return map[string]any{"state": r.State}
}

// OAuth2 Callback
type OAuth2CallbackRequest struct {
	Type         string `json:"type"`
	State        string `json:"state"`
	SessionState string `session:"state"`
	Code         string `json:"code"`
}

type OAuth2CallbackResponse struct {
	RedirectURL string `json:"-"`
	AccessToken string `json:"-"`
}

func (r OAuth2CallbackResponse) RedirectInfo() (int, string) {
	return http.StatusTemporaryRedirect, r.RedirectURL
}

func (r OAuth2CallbackResponse) AccessTokenInfo() string {
	return r.AccessToken
}

// Wallet Login
type WalletLoginRequest struct {
	Address string `json:"address"`
}

type WalletLoginResponse struct {
	Address string `json:"-"`
	Nonce   string `json:"nonce"`
}

func (r WalletLoginResponse) SessionInfo() map[string]any {
	return map[string]any{"address": r.Address, "nonce": r.Nonce}
}

// Wallet Verify
type WalletVerifyRequest struct {
	Signature      string `json:"signature"`
	SessionNonce   string `session:"nonce"`
	SessionAddress string `session:"address"`
}

type WalletVerifyResponse struct {
	AccessToken string `json:"access_token"`
}

func (r WalletVerifyResponse) AccessTokenInfo() string {
	return r.AccessToken
}
