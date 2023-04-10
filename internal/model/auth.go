package model

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
type OAuth2VerifyRequest struct {
	Type        string `json:"type"`
	AccessToken string `json:"access_token"`
}

type OAuth2VerifyResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
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
