package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DB  DBConfig `yaml:"db"`
	Sys System   `yaml:"sys"`
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
