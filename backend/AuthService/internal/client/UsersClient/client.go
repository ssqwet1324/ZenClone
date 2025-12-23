package UsersClient

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

type ClientProvider interface {
	AddUser(ctx context.Context, userData RegisterRequest) error
	CompareAuthData(ctx context.Context, userData AuthRequest) (string, error)
	GetRefreshToken(ctx context.Context, token TokenRequest) (string, error)
	UpdateRefreshToken(ctx context.Context, req UpdateRefreshTokenRequest) error
}

// clientService - структура клиента
type clientService struct {
	client  *resty.Client
	baseUrl string
	log     *zap.Logger
}

// New - конструктор
func New(client *resty.Client, log *zap.Logger, cfg *ConfigUsersServiceClient) ClientProvider {
	client.
		SetRetryCount(int(cfg.RetryCount)).
		SetRetryWaitTime(cfg.RetryDelay)

	return &clientService{
		client:  client,
		baseUrl: cfg.BaseURL,
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
		Post(c.baseUrl + addUser)

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
func (c *clientService) CompareAuthData(ctx context.Context, userData AuthRequest) (string, error) {
	var errResp entity.ErrorResponse

	response, err := c.client.R().
		SetContext(ctx).
		SetBody(userData).
		SetError(&errResp).
		Post(c.baseUrl + compareAuthData)

	if err != nil {
		c.log.Error("Error comparing auth data", zap.Error(err))
		return "", entity.ErrInternalServer
	}

	if response.IsError() {
		return "", errResp
	}

	var authResponse AuthResponse
	if err := json.Unmarshal(response.Body(), &authResponse); err != nil {
		c.log.Error("Error unmarshalling AuthResponse", zap.Error(err))
		return "", entity.ErrCannotParseClaims
	}

	return authResponse.ID, nil
}

// GetRefreshToken  - получаем refresh токен
func (c *clientService) GetRefreshToken(ctx context.Context, token TokenRequest) (string, error) {
	var errResp entity.ErrorResponse

	response, err := c.client.R().
		SetContext(ctx).
		SetBody(token).
		SetError(&errResp).
		Post(c.baseUrl + getRefreshToken)

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
		Post(c.baseUrl + updateRefreshToken)

	if err != nil {
		c.log.Error("Error updating refresh token", zap.Error(err))
		return entity.ErrInternalServer
	}

	if response.IsError() {
		return errResp
	}

	return nil
}
