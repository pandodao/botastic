package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v3"
)

type IndexStoreDriver string

const (
	IndexStoreMemory IndexStoreDriver = "memory"
	IndexStoreMilvus IndexStoreDriver = "milvus"
	IndexStoreRedis  IndexStoreDriver = "redis"
)

type Mode string

const (
	ModeSaaS  Mode = "saas"
	ModeLocal Mode = "local"
)

type Config struct {
	Mode       Mode             `yaml:"mode"`
	SaaS       SaasModeConfig   `yaml:"saas"`
	Local      LocalModeConfig  `yaml:"local"`
	DB         DBConfig         `yaml:"db"`
	Sys        System           `yaml:"sys"`
	OpenAI     OpenAIConfig     `yaml:"openai"`
	IndexStore IndexStoreConfig `yaml:"index_store"`
}

type IndexStoreConfig struct {
	Driver    IndexStoreDriver        `yaml:"driver"`
	Dimension int                     `yaml:"dimension"`
	Milvus    *IndexStoreMilvusConfig `yaml:"milvus"`
	Redis     *IndexStoreRedisConfig  `yaml:"redis"`
}

type OrderSyncerConfig struct {
	Interval       time.Duration `yaml:"interval"`
	CheckInterval  time.Duration `yaml:"check_interval"`
	CancelInterval time.Duration `yaml:"cancel_interval"`
}

type Mixpay struct {
	PayeeId           string `yaml:"payee_id"`
	QuoteAssetId      string `yaml:"quote_asset_id"`
	SettlementAssetId string `yaml:"settlement_asset_id"`
	CallbackUrl       string `yaml:"callback_url"`
	ReturnTo          string `yaml:"return_to"`
	FailedReturnTo    string `yaml:"failed_return_to"`
}

type IndexStoreMilvusConfig struct {
	Collection          string `yaml:"collection"`
	CollectionShardsNum int32  `yaml:"collection_shards_num"`
	Address             string `yaml:"address"`
}

type IndexStoreRedisConfig struct {
	Address   string `yaml:"address"`
	Password  string `yaml:"password"`
	DB        int    `yaml:"db"`
	KeyPrefix string `yaml:"key_prefix"`
}

type Lemonsqueezy struct {
	Key     string `yaml:"key"`
	StoreID int64  `yaml:"store_id"`
}

type Twitter struct {
	ApiKey      string `yaml:"api_key"`
	ApiSecret   string `yaml:"api_secret"`
	CallbackUrl string `yaml:"callback_url"`
}

type TopupVariant struct {
	Name    string  `yaml:"name" json:"name"`
	Amount  float64 `yaml:"amount" json:"amount"`
	LemonID int64   `yaml:"lemon_id" json:"lemon_id"`
}

type OpenAIConfig struct {
	Keys    []string      `yaml:"keys"`
	Timeout time.Duration `yaml:"timeout"`
}

type DBConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type System struct {
	SecretKey string `yaml:"secret_key"`
}

type LocalModeConfig struct{}

type SaasModeConfig struct {
	// how many apps and bots a user can create, zero means unlimited
	AppPerUserLimit int `yaml:"app_per_user_limit"`
	BotPerUserLimit int `yaml:"bot_per_user_limit"`

	ExtraRate       float64           `yaml:"extra_rate"`
	InitUserCredits float64           `yaml:"init_user_credits"`
	Auth            Auth              `yaml:"auth"`
	Mixpay          Mixpay            `yaml:"mixpay"`
	OrderSyncer     OrderSyncerConfig `yaml:"order_syncer"`
	Lemon           Lemonsqueezy      `yaml:"lemonsqueezy"`
	TopupVariants   []TopupVariant    `yaml:"topup_variants"`
	Twitter         Twitter           `yaml:"twitter"`
}

type Auth struct {
	JwtSecret         string   `yaml:"jwt_secret"`
	MixinClientSecret string   `yaml:"mixin_client_secret"`
	TrustDomains      []string `yaml:"trust_domains"`
}

func (c Config) validate() error {
	switch c.IndexStore.Driver {
	case IndexStoreMemory:
	case IndexStoreMilvus:
		if c.IndexStore.Milvus == nil {
			return fmt.Errorf("index_store.milvus is required")
		}
	case IndexStoreRedis:
		if c.IndexStore.Redis == nil {
			return fmt.Errorf("index_store.redis is required")
		}
	default:
		return fmt.Errorf("index_store.driver is invalid: %s", c.IndexStore.Driver)
	}

	switch c.Mode {
	case ModeSaaS, ModeLocal:
	default:
		return errors.New("invalid mode")
	}
	return nil
}

func DefaultConfigString() string {
	cfg := defaultConfig()
	data, _ := yaml.Marshal(cfg)
	return string(data)
}

func defaultConfig() *Config {
	return &Config{
		Mode: ModeLocal,
		SaaS: SaasModeConfig{
			AppPerUserLimit: 10,
			BotPerUserLimit: 10,
			OrderSyncer: OrderSyncerConfig{
				Interval:       time.Second,
				CheckInterval:  10 * time.Second,
				CancelInterval: 2 * time.Hour,
			},
		},
		Sys: System{},
		IndexStore: IndexStoreConfig{
			Driver: IndexStoreMemory,
		},
	}
}

var cfg *Config

func C() *Config {
	if cfg == nil {
		cfg = MustInit("./config.yaml")
	}
	return cfg
}

func MustInit(fp string) *Config {
	cfg, err := Init(fp)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func Init(fp string) (*Config, error) {
	c := defaultConfig()
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	return c, nil
}
