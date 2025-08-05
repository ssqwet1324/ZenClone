package PostClient

import "time"

type PostServiceClient struct {
	BaseURL    string
	RetryCount int64
	RetryDelay time.Duration
}
