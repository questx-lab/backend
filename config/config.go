package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Configs struct {
	Env string

	Database        DatabaseConfigs
	ApiServer       APIServerConfigs
	GameProxyServer ServerConfigs
	Auth            AuthConfigs
	Session         SessionConfigs
	Storage         S3Configs
	File            FileConfigs
	Quest           QuestConfigs
	Redis           RedisConfigs
	Kafka           KafkaConfigs
	Game            GameConfigs
	SearchServer    SearchServerConfigs
	Eth             EthConfigs
}

type DatabaseConfigs struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
	LogLevel string
}

func (d DatabaseConfigs) ConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
	)
}

type ServerConfigs struct {
	Host      string
	Port      string
	AllowCORS []string
}

func (c ServerConfigs) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type APIServerConfigs struct {
	ServerConfigs
	MaxLimit     int
	DefaultLimit int
}

type SessionConfigs struct {
	Secret string
	Name   string
}

type AuthConfigs struct {
	TokenSecret  string
	AccessToken  TokenConfigs
	RefreshToken TokenConfigs

	Google   OAuth2Config
	Twitter  OAuth2Config
	Discord  OAuth2Config
	Telegram TelegramConfigs
}

type OAuth2Config struct {
	Name      string
	VerifyURL string
	IDField   string

	// Only for verifying id token or authorization code. Leave as empty to
	// disable this feature. The issuer must follow OpenID interface.
	Issuer string
	// Use this field to verify authorization code in case the issuer doesn't
	// follow OpenID interface.
	// NOTE: This field cannot be used to verify id token.
	TokenURL string
	ClientID string
}

type TelegramConfigs struct {
	ReclaimDelay time.Duration

	Name            string
	BotToken        string
	LoginExpiration time.Duration
}

type TokenConfigs struct {
	Name       string
	Expiration time.Duration
}

type FileConfigs struct {
	MaxSize int64

	AvatarCropHeight uint
	AvatarCropWidth  uint
}

type TwitterConfigs struct {
	ReclaimDelay time.Duration

	AppAccessToken string

	ConsumerAPIKey    string
	ConsumerAPISecret string
	AccessToken       string
	AccessTokenSecret string
}

type DiscordConfigs struct {
	ReclaimDelay time.Duration

	BotToken string
	BotID    string
}

type QuestConfigs struct {
	Twitter  TwitterConfigs
	Dicord   DiscordConfigs
	Telegram TelegramConfigs

	QuizMaxQuestions                 int
	QuizMaxOptions                   int
	InviteReclaimDelay               time.Duration
	InviteCommunityReclaimDelay      time.Duration
	InviteCommunityRequiredFollowers int
	InviteCommunityRewardToken       string
	InviteCommunityRewardAmount      float64
}

type RedisConfigs struct {
	Addr string
}

type KafkaConfigs struct {
	Addr string
}

type GameConfigs struct {
	GameSaveFrequency time.Duration

	MoveActionDelay time.Duration
	InitActionDelay time.Duration
	JoinActionDelay time.Duration
}

type SearchServerConfigs struct {
	ServerConfigs

	RPCName  string
	IndexDir string

	SearchServerEndpoint string
}

type S3Configs struct {
	Region   string
	Endpoint string

	AccessKey   string
	SecretKey   string
	SSLDisabled bool
}

type EthConfigs struct {
	Chains map[string]ChainConfig `toml:"chains"`
	Keys   KeyConfigs
}

type ChainConfig struct {
	Chain string   `toml:"chain" json:"chain"`
	Rpcs  []string `toml:"rpcs" json:"rpcs"`
	Wss   []string `toml:"wss" json:"wss"`

	// ETH
	UseEip1559 bool `toml:"use_eip_1559" json:"use_eip_1559"` // For gas calculation

	BlockTime  int `toml:"block_time"`
	AdjustTime int `toml:"adjust_time"`

	ThresholdUpdateBlock int `toml:"threshold_update_block"`
}

func LoadEthConfigs() EthConfigs {
	tomlFile := "./chain.toml"
	if _, err := os.Stat(tomlFile); os.IsNotExist(err) {
		panic(err)
	}
	var cfg EthConfigs
	_, err := toml.DecodeFile(tomlFile, &cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}

type KeyConfigs struct {
	PubKey  string
	PrivKey string
}
