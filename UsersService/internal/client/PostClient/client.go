package PostClient

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strings"
)

const (
	getUserPosts = "/posts/by-user/"
)

type ClientProvider interface {
	GetPostsUser(ctx context.Context, userID uuid.UUID) ([]PostResponse, error)
}

type clientService struct {
	client  *resty.Client
	log     *zap.Logger
	baseUrl string
}

func New(client *resty.Client, log *zap.Logger, cfg *PostServiceClient) ClientProvider {
	client.
		SetRetryCount(int(cfg.RetryCount)).
		SetRetryWaitTime(cfg.RetryDelay)

	return &clientService{
		client:  client,
		log:     log.Named("usersClient"),
		baseUrl: cfg.BaseURL,
	}
}

// GetPostsUser - получаем посты пользователя для профиля по его id
func (c *clientService) GetPostsUser(ctx context.Context, userID uuid.UUID) ([]PostResponse, error) {
	var postList PostListResponse

	url := fmt.Sprintf("%s%s%s", strings.TrimRight(c.baseUrl, "/"), getUserPosts, userID.String())

	response, err := c.client.R().
		SetContext(ctx).
		SetResult(&postList).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("GetPostsUser: error sending request: %w", err)
	}

	if response.IsError() {
		return nil, fmt.Errorf("GetPostsUser: status %d, body: %s", response.StatusCode(), response.String())
	}

	return postList.Posts, nil
}
