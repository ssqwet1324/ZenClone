package UsersClient

import "time"

type ConfigUsersServiceClient struct {
	BaseURL    string
	RetryCount int64
	RetryDelay time.Duration
}
