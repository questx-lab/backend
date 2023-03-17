package models

type Auth struct {
	UserName string
	Password string

	GmailToken    string
	FacebookToken string
	TwitterToken  string
}

type Token struct {
}
