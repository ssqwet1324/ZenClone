package subsclient

import (
	"PostService/internal/entity"
	"context"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	userServiceURL           = "/api/v1/user/subs/{username}"
	userProfileByUsernameURL = "/api/v1/get-user-profile/{username}"
)

// ClientProvider - интерфейс клиента
type ClientProvider interface {
	GetSubsUser(ctx context.Context, username string, accessToken string) (*SubsList, error)
	GetUserProfileByUsername(ctx context.Context, username string, accessToken string) (*UserProfile, error)
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
		log:     log.Named("SubClient"),
	}
}

// GetSubsUser - получаем подписки пользователя
func (c *clientService) GetSubsUser(ctx context.Context, username string, accessToken string) (*SubsList, error) {
	var errResp entity.ErrorResponse
	var subsList SubsList

	url := c.baseURL + userServiceURL

	response, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetPathParam("username", username).
		SetResult(&subsList).
		SetError(&errResp).
		Get(url)
	if err != nil {
		c.log.Error("Error getting subs user - network error",
			zap.String("username", username),
			zap.String("url", url),
			zap.Error(err))
		return nil, entity.ErrInternalError
	}

	if response.IsError() {
		c.log.Error("Error getting subs user - HTTP error",
			zap.String("username", username),
			zap.Int("status_code", response.StatusCode()),
			zap.String("error_code", errResp.ErrorDetail.Code),
			zap.String("error_message", errResp.ErrorDetail.Message),
			zap.String("response_body", string(response.Body())))
		return nil, errResp
	}

	return &subsList, nil
}

// GetUserProfileByUsername - получаем профиль пользователя по username
func (c *clientService) GetUserProfileByUsername(ctx context.Context, username string, accessToken string) (*UserProfile, error) {
	var errResp entity.ErrorResponse
	var profile UserProfile

	url := c.baseURL + userProfileByUsernameURL

	response, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetPathParam("username", username).
		SetResult(&profile).
		SetError(&errResp).
		Get(url)
	if err != nil {
		c.log.Error("Error getting user profile - network error",
			zap.String("username", username),
			zap.String("url", url),
			zap.Error(err))
		return nil, entity.ErrInternalError
	}

	if response.IsError() {
		c.log.Error("Error getting user profile - HTTP error",
			zap.String("username", username),
			zap.Int("status_code", response.StatusCode()),
			zap.String("error_code", errResp.ErrorDetail.Code),
			zap.String("error_message", errResp.ErrorDetail.Message))
		return nil, errResp
	}

	return &profile, nil
}
