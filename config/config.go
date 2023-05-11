package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
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

type StateConfig struct {
	WorkerCount int `yaml:"worker_count"`
}

type LLMsConfig struct {
	Enabled []string             `yaml:"enabled"`
	Items   map[string]LLMConfig `yaml:"items"`
}

type LLMConfig struct {
	Provider LLMProvider   `yaml:"provider"`
	OpenAI   *OpenAIConfig `yaml:"openai,omitempty"`
}

type OpenAIConfig struct {
	Key                     string        `yaml:"key"`
	ChatModels              []string      `yaml:"chat_models"`
	EmbeddingModels         []string      `yaml:"embedding_models"`
	ChatRequestTimeout      time.Duration `yaml:"chat_request_timeout"`
	EmbeddingRequestTimeout time.Duration `yaml:"embedding_request_timeout"`
}

type HttpdConfig struct {
	Debug bool   `yaml:"debug"`
	Addr  string `yaml:"addr"`
}

type VectorStorageConfig struct {
	Driver VectorStorageDriver        `yaml:"driver"`
	Milvus *VectorStorageMilvusConfig `yaml:"milvus,omitempty"`
	Redis  *VectorStorageRedisConfig  `yaml:"redis,omitempty"`
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

func (c Config) validate() error {
	switch c.VectorStorage.Driver {
	case VectorStorageMemory:
	case VectorStorageMilvus:
		if c.VectorStorage.Milvus == nil {
			return fmt.Errorf("vector_storage.milvus is required")
		}
	case VectorStorageRedis:
		if c.VectorStorage.Redis == nil {
			return fmt.Errorf("vector_storage.redis is required")
		}
	default:
		return fmt.Errorf("vector_storage.driver is invalid: %s", c.VectorStorage.Driver)
	}

	switch c.DB.Driver {
	case DBSqlite, DBMysql, DBPostgres:
	default:
		return fmt.Errorf("db.driver is invalid: %s", c.DB.Driver)
	}
	if c.DB.DSN == "" {
		return fmt.Errorf("db.dsn is required")
	}

	for _, name := range c.LLMs.Enabled {
		v, ok := c.LLMs.Items[name]
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

func ExampleConfig() *Config {
	return &Config{
		Httpd: HttpdConfig{
			Debug: true,
			Addr:  ":8080",
		},
		DB: DBConfig{
			Driver: "sqlite",
			DSN:    "file::memory:?cache=shared",
		},
		VectorStorage: VectorStorageConfig{
			Driver: VectorStorageMemory,
		},
		LLMs: LLMsConfig{
			Enabled: []string{"openai-1"},
			Items: map[string]LLMConfig{
				"openai-1": {
					Provider: LLMProviderOpenAI,
					OpenAI: &OpenAIConfig{
						Key:                     "YOUR_OPENAI_KEY",
						ChatModels:              []string{"gpt-3.5-turbo", "gpt-4"},
						EmbeddingModels:         []string{"text-embedding-ada-002"},
						ChatRequestTimeout:      8 * time.Second,
						EmbeddingRequestTimeout: 10 * time.Second,
					},
				},
			},
		},
	}
}

func DefaultConfig() *Config {
	return &Config{
		Httpd: HttpdConfig{
			Debug: true,
			Addr:  ":8080",
		},
		DB: DBConfig{
			Driver: "sqlite",
			DSN:    "file::memory:?cache=shared",
		},
		VectorStorage: VectorStorageConfig{
			Driver: VectorStorageMemory,
		},
	}
}

func MustInit(fp string) *Config {
	cfg, err := Init(fp)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
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
