package model

// OAuth2
type OAuth2VerifyRequest struct {
	Type        string `json:"type"`
	AccessToken string `json:"access_token"`
}

type OAuth2VerifyResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type OAuth2IDVerifyRequest struct {
	Type    string `json:"type"`
	IDToken string `json:"id_token"`
}

type OAuth2IDVerifyResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type OAuth2LinkRequest struct {
	Type        string `json:"type"`
	AccessToken string `json:"access_token"`
}

type OAuth2LinkResponse struct{}

// Wallet
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

type WalletVerifyRequest struct {
	Signature      string `json:"signature"`
	SessionNonce   string `session:"nonce,delete"`
	SessionAddress string `session:"address,delete"`
}

type WalletVerifyResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type WalletLinkRequest struct {
	Signature      string `json:"signature"`
	SessionNonce   string `session:"nonce,delete"`
	SessionAddress string `session:"address,delete"`
}

type WalletLinkResponse struct{}

// Telegram
type TelegramLinkRequest struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int    `json:"auth_date"`
	Hash      string `json:"hash"`
}

type TelegramLinkResponse struct{}

// Refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
