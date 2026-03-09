package usersclient

import (
	"AuthService/internal/entity"
	"context"
	"encoding/json"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	addUser            = "/internal/add-user"
	compareAuthData    = "/internal/compare-auth-data"
	getRefreshToken    = "/internal/get-refresh-token"
	updateRefreshToken = "/internal/update-refresh-token"
)

// ClientProvider - интерфейс авторизации
type ClientProvider interface {
	AddUser(ctx context.Context, userData RegisterRequest) error
	CompareAuthData(ctx context.Context, userData AuthRequest) (*AuthResponse, error)
	GetRefreshToken(ctx context.Context, token TokenRequest) (string, error)
	UpdateRefreshToken(ctx context.Context, req UpdateRefreshTokenRequest) error
}

// clientService - структура клиента
type clientService struct {
	client  *resty.Client
	baseURL string
	log     *zap.Logger
}

// New - конструктор
func New(client *resty.Client, log *zap.Logger, cfg *ConfigUsersServiceClient) ClientProvider {
	client.
		SetRetryCount(int(cfg.RetryCount)).
		SetRetryWaitTime(cfg.RetryDelay)

	return &clientService{
		client:  client,
		baseURL: cfg.BaseURL,
		log:     log.Named("AuthClient"),
	}
}

// AddUser - добавление пользователя
func (c *clientService) AddUser(ctx context.Context, userData RegisterRequest) error {
	var errResp entity.ErrorResponse

	response, err := c.client.R().
		SetContext(ctx).
		SetBody(userData).
		SetError(&errResp).
		Post(c.baseURL + addUser)

	if err != nil {
		c.log.Error("Error adding user", zap.Error(err))
		return entity.ErrInternalServer
	}

	if response.IsError() {
		return errResp
	}

	return nil
}

// CompareAuthData - сравнение данных для входа
func (c *clientService) CompareAuthData(ctx context.Context, userData AuthRequest) (*AuthResponse, error) {
	var errResp entity.ErrorResponse

	response, err := c.client.R().
		SetContext(ctx).
		SetBody(userData).
		SetError(&errResp).
		Post(c.baseURL + compareAuthData)

	if err != nil {
		c.log.Error("Error comparing auth data", zap.Error(err))
		return nil, entity.ErrInternalServer
	}

	if response.IsError() {
		return nil, errResp
	}

	var authResponse AuthResponse
	if err := json.Unmarshal(response.Body(), &authResponse); err != nil {
		c.log.Error("Error unmarshalling AuthResponse", zap.Error(err))
		return nil, entity.ErrCannotParseClaims
	}

	return &authResponse, nil
}

// GetRefreshToken  - получаем refresh токен
func (c *clientService) GetRefreshToken(ctx context.Context, token TokenRequest) (string, error) {
	var errResp entity.ErrorResponse

	response, err := c.client.R().
		SetContext(ctx).
		SetBody(token).
		SetError(&errResp).
		Post(c.baseURL + getRefreshToken)

	if err != nil {
		return "", entity.ErrInternalServer
	}

	if response.IsError() {
		return "", errResp
	}

	var tokenResponse TokenResponse
	if err := json.Unmarshal(response.Body(), &tokenResponse); err != nil {
		c.log.Error("Error unmarshalling TokenResponse", zap.Error(err))
		return "", entity.ErrCannotParseClaims
	}

	return tokenResponse.RefreshToken, nil
}

// UpdateRefreshToken - обновляем токен
func (c *clientService) UpdateRefreshToken(ctx context.Context, req UpdateRefreshTokenRequest) error {
	var errResp entity.ErrorResponse

	response, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetError(&errResp).
		Post(c.baseURL + updateRefreshToken)

	if err != nil {
		c.log.Error("Error updating refresh token", zap.Error(err))
		return entity.ErrInternalServer
	}

	if response.IsError() {
		return errResp
	}

	return nil
}
