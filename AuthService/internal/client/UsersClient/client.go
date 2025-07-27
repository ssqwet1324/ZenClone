package UsersClient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	addUser            = "/add-user"
	compareAuthData    = "/compare-auth-data"
	getRefreshToken    = "/get-refresh-token"
	updateRefreshToken = "/update-refresh-token"
)

type ClientProvider interface {
	AddUser(ctx context.Context, userData RegisterRequest) error
	CompareAuthData(ctx context.Context, userData AuthRequest) (string, error)
	GetRefreshToken(ctx context.Context, token TokenRequest) (string, error)
	UpdateRefreshToken(ctx context.Context, req UpdateRefreshTokenRequest) error
}

type clientService struct {
	client  *resty.Client
	baseUrl string
	log     *zap.Logger
}

func New(client *resty.Client, log *zap.Logger, cfg *ConfigUsersServiceClient) ClientProvider {
	client.
		SetRetryCount(int(cfg.RetryCount)).
		SetRetryWaitTime(cfg.RetryDelay)

	return &clientService{
		client:  client,
		baseUrl: cfg.BaseURL,
		log:     log.Named("accountServiceClient"),
	}
}

func (c *clientService) AddUser(ctx context.Context, userData RegisterRequest) error {
	response, err := c.client.R().SetContext(ctx).SetBody(userData).Post(c.baseUrl + addUser)

	if err != nil {
		return fmt.Errorf("AddUser: %w", err)
	}

	if response.IsError() {
		return fmt.Errorf("AddUser: %s", response.Body())
	}

	return nil
}

func (c *clientService) CompareAuthData(ctx context.Context, userData AuthRequest) (string, error) {
	var authResponse AuthResponse
	response, err := c.client.R().SetContext(ctx).SetBody(userData).Post(c.baseUrl + compareAuthData)

	if err != nil {
		return "", fmt.Errorf("CompareAuthData: %w", err)
	}

	if response.IsError() {
		return "", fmt.Errorf("CompareAuthData: %s", response.Body())
	}

	fmt.Println("AuthResponse client:", string(response.Body()))

	if err := json.Unmarshal(response.Body(), &authResponse); err != nil {
		return "", fmt.Errorf("CompareAuthData: failed to parse response: %w", err)
	}

	return authResponse.ID, nil
}

func (c *clientService) GetRefreshToken(ctx context.Context, token TokenRequest) (string, error) {
	var tokenResponse TokenResponse
	response, err := c.client.R().SetContext(ctx).SetBody(token).Post(c.baseUrl + getRefreshToken)
	if err != nil {
		return "", fmt.Errorf("GetRefreshToken: %w", err)
	}

	if response.IsError() {
		return "", fmt.Errorf("GetRefreshToken: %s", response.Body())
	}

	if err := json.Unmarshal(response.Body(), &tokenResponse); err != nil {
		return "", fmt.Errorf("GetRefreshToken: failed to parse response: %w", err)
	}

	return tokenResponse.RefreshToken, nil
}

func (c *clientService) UpdateRefreshToken(ctx context.Context, req UpdateRefreshTokenRequest) error {
	response, err := c.client.R().SetContext(ctx).SetBody(req).Post(c.baseUrl + updateRefreshToken)
	if err != nil {
		return fmt.Errorf("UpdateRefreshToken: %w", err)
	}
	if response.IsError() {
		return fmt.Errorf("UpdateRefreshToken: %s", response.Body())
	}

	return nil
}
