package subsclient

import "time"

// ConfigUsersServiceClient - конфигурация клиента
type ConfigUsersServiceClient struct {
	BaseURL    string        `env:"BASE_URL"`
	RetryCount int64         `env:"RETRY_COUNT"`
	RetryDelay time.Duration `env:"RETRY_DELAY"`
}
