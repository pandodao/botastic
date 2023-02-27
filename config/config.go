package config

import (
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DB       DBConfig       `yaml:"db"`
	Milvus   Milvus         `yaml:"milvus"`
	Sys      System         `yaml:"sys"`
	OpenAPI  OpenAIConfig   `yaml:"openai"`
	TokenCal TokenCalConfig `yaml:"tokencal"`
}

type Milvus struct {
	Address string `yaml:"address"`
}

type OpenAIConfig struct {
	Keys    []string      `yaml:"keys"`
	Timeout time.Duration `yaml:"timeout"`
}

type TokenCalConfig struct {
	Addr string `yaml:"addr"`
}

type DBConfig struct {
	Driver  string `yaml:"driver"`
	DSN     string `yaml:"dsn"`
	AutoGen bool   `yaml:"auto_gen"`
}

type System struct {
	SecretKey string `yaml:"secret_key"`
}

func (c Config) validate() error {
	return nil
}

func DefaultConfigString() string {
	cfg := defaultConfig()
	data, _ := yaml.Marshal(cfg)
	return string(data)
}

func defaultConfig() *Config {
	return &Config{}
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
