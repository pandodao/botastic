package config

type VectorStorageDriver string

const (
	VectorStorageDB    VectorStorageDriver = "db"
	VectorStorageRedis VectorStorageDriver = "redis"
)

type DBDriver string

const (
	DBSqlite   DBDriver = "sqlite"
	DBMysql    DBDriver = "mysql"
	DBPostgres DBDriver = "postgres"
)

type LLMProvider string

const (
	LLMProviderOpenAI LLMProvider = "openai"
)
