package twitter

type User struct {
	Name     string `mapstructure:"name"`
	Handle   string `mapstructure:"handle"`
	PhotoURL string `mapstructure:"photo_url"`
}

type Tweet struct {
	ID     string `mapstructure:"id"`
	Author string `mapstructure:"author"`
	Text   string `mapstructure:"text"`
}
