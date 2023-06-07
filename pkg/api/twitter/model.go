package twitter

type User struct {
	ID         string
	Name       string
	ScreenName string
	PhotoURL   string
}

type Tweet struct {
	ReplyToTweetID   string
	ID               string
	AuthorScreenName string
	Text             string
}
