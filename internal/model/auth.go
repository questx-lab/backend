package model

import (
	"net/http"
	"time"

	"github.com/questx-lab/backend/pkg/xcontext"
)

// Access Token and Refresh Token
type AccessToken struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type RefreshToken struct {
	Family  string
	Counter uint64
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
	SessionState string `session:"state,delete"`
	Code         string `json:"code"`
}

type OAuth2CallbackResponse struct {
	RedirectURL  string `json:"-"`
	AccessToken  string `json:"-"`
	RefreshToken string `json:"-"`
}

func (r OAuth2CallbackResponse) RedirectInfo() (int, string) {
	return http.StatusTemporaryRedirect, r.RedirectURL
}

func (r OAuth2CallbackResponse) CookieInfo(ctx xcontext.Context) []http.Cookie {
	return []http.Cookie{
		{
			Name:     ctx.Configs().Auth.AccessToken.Name,
			Value:    r.AccessToken,
			Path:     "/",
			Domain:   "",
			Expires:  time.Now().Add(ctx.Configs().Auth.AccessToken.Expiration),
			Secure:   true,
			HttpOnly: false,
		},
		{
			Name:     ctx.Configs().Auth.RefreshToken.Name,
			Value:    r.RefreshToken,
			Path:     "/",
			Domain:   "",
			Expires:  time.Now().Add(ctx.Configs().Auth.RefreshToken.Expiration),
			Secure:   true,
			HttpOnly: false,
		},
	}
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
	SessionNonce   string `session:"nonce,delete"`
	SessionAddress string `session:"address,delete"`
}

type WalletVerifyResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
