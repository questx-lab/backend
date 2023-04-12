package twitter

type User struct {
	ID         string `mapstructure:"id_str"`
	Name       string `mapstructure:"name"`
	ScreenName string `mapstructure:"screen_name"`
}

type Tweet struct {
	ReplyToTweetID   string
	ID               string
	AuthorScreenName string
	Text             string
}
