package config

import (
	"fmt"
	"time"

	"github.com/questx-lab/backend/pkg/storage"
)

type Configs struct {
	Env string

	Database      DatabaseConfigs
	ApiServer     ServerConfigs
	WsProxyServer ServerConfigs
	Auth          AuthConfigs
	Session       SessionConfigs
	Storage       storage.S3Configs
	File          FileConfigs
	Quest         QuestConfigs
	Redis         RedisConfigs
	Kafka         KafkaConfigs
	Eth           EthConfigs
}

type DatabaseConfigs struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

func (d *DatabaseConfigs) ConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
	)
}

type ServerConfigs struct {
	Host string
	Port string
	Cert string
	Key  string
}

type SessionConfigs struct {
	Secret string
	Name   string
}

type AuthConfigs struct {
	TokenSecret  string
	AccessToken  TokenConfigs
	RefreshToken TokenConfigs

	Google  OAuth2Config
	Twitter OAuth2Config
}

type OAuth2Config struct {
	Name      string
	VerifyURL string
	IDField   string
}

type TokenConfigs struct {
	Name       string
	Expiration time.Duration
}

type FileConfigs struct {
	MaxSize int
}

type TwitterConfigs struct {
	ReclaimDelay time.Duration

	AppAccessToken string

	ConsumerAPIKey    string
	ConsumerAPISecret string
	AccessToken       string
	AccessTokenSecret string
}

type QuestConfigs struct {
	Twitter TwitterConfigs
}

type RedisConfigs struct {
	Addr string
}

type KafkaConfigs struct {
	Addr string
}

type EthConfigs struct {
	chains map[string]ChainConfig
}

type ChainConfig struct {
	Chain string   `toml:"chain" json:"chain"`
	Rpcs  []string `toml:"rpcs" json:"rpcs"`
	Wss   []string `toml:"wss" json:"wss"`

	// ETH
	UseEip1559 bool `toml:"use_eip_1559" json:"use_eip_1559"` // For gas calculation

	BlockTime  int
	AdjustTime int
}
