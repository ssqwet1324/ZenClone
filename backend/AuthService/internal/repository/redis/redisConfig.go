package redis

// Config  - конфиг redis
type Config struct {
	RedisAddr string `env:"REDIS_ADDR"`
	RedisPWD  string `env:"REDIS_PWD"`
	RedisDB   int64  `env:"REDIS_DB"`
}
