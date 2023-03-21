package model

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
}

// OAuth Callback
type OAuth2CallbackRequest struct {
	Type  string `json:"type"`
	State string `json:"state"`
	Code  string `json:"code"`
}

type OAuth2CallbackResponse struct {
}

// Wallet Login
type WalletLoginRequest struct {
	Address string `json:"address"`
}

type WalletLoginResponse struct {
	Nonce string `json:"nonce"`
}

// Wallet Verify
type WalletVerifyRequest struct {
	Signature string `json:"signature"`
}

type WalletVerifyResponse struct {
	AccessToken string `json:"access_token"`
}
