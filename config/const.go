package config

type VectorStorageDriver string

const (
	VectorStorageMemory VectorStorageDriver = "memory"
	VectorStorageMilvus VectorStorageDriver = "milvus"
	VectorStorageRedis  VectorStorageDriver = "redis"
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
