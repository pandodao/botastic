package config

import (
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerAddress string   `yaml:"server_address"`
	DB            DBConfig `yaml:"db"`
	Dapp          Dapp     `yaml:"dapp"`
	System        System   `yaml:"system"`
	Admins        []string `yaml:"admins"`
	Badger        Badger   `yaml:"badger"`
	Group         Group    `yaml:"group"`
	MtgHub        MtgHub   `yaml:"mtghub"`
}

type MtgHub struct {
	CashierBatch    int  `yaml:"cashier_batch"`
	CashierCapacity int  `yaml:"cashier_capacity"`
	Replay          bool `yaml:"replay"`
}

type DBConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type Badger struct {
	Path       string        `yaml:"path"`
	GcDuration time.Duration `yaml:"gc_duration"`
}

type Group struct {
	Members   []string `yaml:"members"`
	Threshold uint8    `yaml:"threshold"`
}

type System struct {
	Lang              string  `json:"lang"`
	SecretKey         string  `json:"secret_key"`
	BonusPercent      float64 `json:"bonus_percent"`
	FswapSlippage     float64 `json:"4swap_slippage"`
	ExchangeAccountID string  `json:"exchange_account_id"`
	ProfitAccountID   string  `json:"profit_account_id"`
}

func (c Config) validate() error {
	return nil
}

type Dapp struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	SessionID    string `yaml:"session_id"`
	PrivateKey   string `yaml:"private_key"`
	PinToken     string `yaml:"pin_token"`
	Pin          string `yaml:"pin"`
	Scope        string `yaml:"scope"`
}

func DefaultConfigString() string {
	cfg := defaultConfig()
	data, _ := yaml.Marshal(cfg)
	return string(data)
}

func defaultConfig() *Config {
	return &Config{
		ServerAddress: ":7080",
		MtgHub: MtgHub{
			CashierBatch:    100,
			CashierCapacity: 10,
		},
	}
}

var cfg *Config

func C() *Config {
	return cfg
}

func MustInit(fp string) *Config {
	var err error
	cfg, err = Init(fp)
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
