package UsersClient

import "time"

// ConfigUsersServiceClient - конфигурация клиента
type ConfigUsersServiceClient struct {
	BaseURL    string
	RetryCount int64
	RetryDelay time.Duration
}
