package config

type Storage string

var (
	StoragePostgres Storage = "postgres"
)

type PostgresConfig struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	ConnURL  string `json:"conn_url"`
}

type Queue string

var (
	RedisPubSub Queue = "redis"
)

type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DB       int    `json:"db"`
	Password string `json:"password"`
}

type Config struct {
	Host string `json:"host"`
	Port int    `config:"port,backend=flags,short=p"`

	StorageBackend Storage `json:"storage"`
	QueueBackend   Queue   `json:"queue"`

	Postgres PostgresConfig `json:"postgres"`
	Redis    RedisConfig    `json:"redis"`
}
