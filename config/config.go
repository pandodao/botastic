package config

import (
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

type Config struct {
	DB          DBConfig          `yaml:"db"`
	IndexStore  IndexStoreConfig  `yaml:"index_store"`
	Sys         System            `yaml:"sys"`
	OpenAI      OpenAIConfig      `yaml:"openai"`
	Auth        Auth              `yaml:"auth"`
	Mixpay      Mixpay            `yaml:"mixpay"`
	OrderSyncer OrderSyncerConfig `yaml:"order_syncer"`
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

type OpenAIConfig struct {
	Keys    []string      `yaml:"keys"`
	Timeout time.Duration `yaml:"timeout"`
}

type DBConfig struct {
	Driver  string `yaml:"driver"`
	DSN     string `yaml:"dsn"`
	AutoGen bool   `yaml:"auto_gen"`
}

type System struct {
	InitUserCredits float64 `yaml:"init_user_credits"`
	SecretKey       string  `yaml:"secret_key"`
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
	return nil
}

func DefaultConfigString() string {
	cfg := defaultConfig()
	data, _ := yaml.Marshal(cfg)
	return string(data)
}

func defaultConfig() *Config {
	return &Config{
		OrderSyncer: OrderSyncerConfig{
			Interval:       time.Second,
			CheckInterval:  10 * time.Second,
			CancelInterval: 2 * time.Hour,
		},
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
