package config

import (
	"fmt"
	"io/ioutil"

	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Log           LogConfig           `yaml:"log"`
	Httpd         HttpdConfig         `yaml:"httpd"`
	DB            DBConfig            `yaml:"db"`
	VectorStorage VectorStorageConfig `yaml:"vector_storage"`
	LLMs          LLMsConfig          `yaml:"llms"`
	State         StateConfig         `yaml:"state"`
}

func (c Config) String() string {
	data, _ := yaml.Marshal(c)
	return string(data)
}

type LogConfig struct {
	Level string `yaml:"level"`
}

func (c LogConfig) Validate() error {
	_, err := zapcore.ParseLevel(c.Level)
	if err != nil {
		return fmt.Errorf("invalid log.level: %w", err)
	}
	return nil
}

type StateConfig struct {
	WorkerCount int `yaml:"worker_count"`
}

type LLMsConfig struct {
	Enabled []string             `yaml:"enabled"`
	Items   map[string]LLMConfig `yaml:"items"`
}

func (c LLMsConfig) Validate() error {
	for _, name := range c.Enabled {
		v, ok := c.Items[name]
		if !ok {
			return fmt.Errorf("llms.items.%s is required", name)
		}
		switch v.Provider {
		case LLMProviderOpenAI:
			if v.OpenAI == nil {
				return fmt.Errorf("llms.items.%s.openai is required", name)
			}
			for _, m := range v.OpenAI.ChatModels {
				switch m {
				case "gpt-3.5-turbo", "gpt-3.5-turbo-0301", "gpt-4", "gpt-4-0314", "gpt-4-32k", "gpt-4-32k-0314":
				default:
					return fmt.Errorf("llms.items.%s.chat_models is invalid: %s", name, m)
				}
			}
			for _, m := range v.OpenAI.EmbeddingModels {
				switch m {
				case "text-embedding-ada-002":
				default:
					return fmt.Errorf("llms.items.%s.embedding_models is invalid: %s", name, m)
				}
			}

		default:
			return fmt.Errorf("llms.items.%s.provider is invalid: %s", name, v.Provider)
		}
	}
	return nil
}

type LLMConfig struct {
	Provider LLMProvider   `yaml:"provider"`
	OpenAI   *OpenAIConfig `yaml:"openai,omitempty"`
}

type OpenAIConfig struct {
	Key             string   `yaml:"key"`
	ChatModels      []string `yaml:"chat_models"`
	EmbeddingModels []string `yaml:"embedding_models"`
}

type HttpdConfig struct {
	Debug bool   `yaml:"debug"`
	Addr  string `yaml:"addr"`
}

type VectorStorageConfig struct {
	Driver VectorStorageDriver       `yaml:"driver"`
	Redis  *VectorStorageRedisConfig `yaml:"redis,omitempty"`
}

func (c VectorStorageConfig) Validate() error {
	switch c.Driver {
	case VectorStorageDB:
	case VectorStorageRedis:
		if c.Redis == nil {
			return fmt.Errorf("vector_storage.redis is required")
		}
	default:
		return fmt.Errorf("vector_storage.driver is invalid: %s", c.Driver)
	}
	return nil
}

type VectorStorageMilvusConfig struct {
	Collection          string `yaml:"collection"`
	CollectionShardsNum int32  `yaml:"collection_shards_num"`
	Address             string `yaml:"address"`
}

type VectorStorageRedisConfig struct {
	Address   string `yaml:"address"`
	Password  string `yaml:"password"`
	DB        int    `yaml:"db"`
	KeyPrefix string `yaml:"key_prefix"`
}

type DBConfig struct {
	Driver DBDriver `yaml:"driver"`
	DSN    string   `yaml:"dsn"`
	Debug  bool     `yaml:"debug"`
}

func (c DBConfig) Validate() error {
	switch c.Driver {
	case DBSqlite, DBMysql, DBPostgres:
	default:
		return fmt.Errorf("db.driver is invalid: %s", c.Driver)
	}
	if c.DSN == "" {
		return fmt.Errorf("db.dsn is required")
	}

	return nil
}

func (c Config) validate() error {
	for _, v := range []any{c.Log, c.Httpd, c.DB, c.VectorStorage, c.LLMs, c.State} {
		if vi, ok := v.(interface{ Validate() error }); ok {
			if err := vi.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

func ExampleConfig() *Config {
	return &Config{
		Log: LogConfig{
			Level: "debug",
		},
		Httpd: HttpdConfig{
			Debug: true,
			Addr:  ":8080",
		},
		DB: DBConfig{
			Driver: "sqlite",
			DSN:    "file::memory:?cache=shared",
		},
		VectorStorage: VectorStorageConfig{
			Driver: VectorStorageDB,
		},
		State: StateConfig{
			WorkerCount: 10,
		},
		LLMs: LLMsConfig{
			Enabled: []string{"openai-1"},
			Items: map[string]LLMConfig{
				"openai-1": {
					Provider: LLMProviderOpenAI,
					OpenAI: &OpenAIConfig{
						Key:             "YOUR_OPENAI_KEY",
						ChatModels:      []string{"gpt-3.5-turbo", "gpt-4"},
						EmbeddingModels: []string{"text-embedding-ada-002"},
					},
				},
			},
		},
	}
}

func DefaultConfig() *Config {
	return &Config{
		Log: LogConfig{
			Level: "info",
		},
		Httpd: HttpdConfig{
			Debug: true,
			Addr:  ":8080",
		},
		DB: DBConfig{
			Driver: "sqlite",
			DSN:    "file::memory:?cache=shared",
		},
		VectorStorage: VectorStorageConfig{
			Driver: VectorStorageDB,
		},
		State: StateConfig{
			WorkerCount: 10,
		},
	}
}

func Init(fp string) (*Config, error) {
	c := DefaultConfig()
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
