package usecase

import (
	"AuthService/internal/client/usersclient"
	"AuthService/internal/config"
	"AuthService/internal/entity"
	"AuthService/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

const testSecret = "test-secret"

// newTestUseCase - хелпер для создания UseCase с моками
func newTestUseCase(t *testing.T) (*UseCase, *mocks.MockRepositoryProvider, *mocks.MockClientProvider, *gomock.Controller) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockRepositoryProvider(ctrl)
	mockClient := mocks.NewMockClientProvider(ctrl)

	uc := &UseCase{
		repo:   mockRepo,
		log:    zap.NewNop(),
		client: mockClient,
		cfg:    &config.Config{JWTSecret: testSecret},
	}

	return uc, mockRepo, mockClient, ctrl
}

// generateValidToken - хелпер для генерации валидного JWT токена
func generateValidToken(userID string) string {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testSecret))
	return tokenString
}

// ====== RegisterUser ======

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name      string
		req       entity.RegisterRequest
		mockSetup func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider)
		wantErr   error
	}{
		{
			name: "success",
			req: entity.RegisterRequest{
				Login:     "test1",
				Password:  "Qwerty12",
				Username:  "TestUser",
				FirstName: "Test",
				LastName:  "Testov",
			},
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				client.EXPECT().
					AddUser(gomock.Any(), gomock.Any()).
					Return(nil)
				repo.EXPECT().
					SaveRefreshToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "client AddUser failed - user already exists",
			req: entity.RegisterRequest{
				Login:    "test1",
				Password: "Qwerty12",
			},
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				client.EXPECT().
					AddUser(gomock.Any(), gomock.Any()).
					Return(entity.ErrorResponse{
						ErrorDetail: entity.ErrorDetail{
							Code:    "USER_ALREADY_EXISTS",
							Message: "User already exists",
						},
					})
			},
			wantErr: entity.ErrorResponse{
				ErrorDetail: entity.ErrorDetail{
					Code:    "USER_ALREADY_EXISTS",
					Message: "User already exists",
				},
			},
		},
		{
			name: "save refresh token failed",
			req: entity.RegisterRequest{
				Login:    "test1",
				Password: "Qwerty12",
			},
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				client.EXPECT().
					AddUser(gomock.Any(), gomock.Any()).
					Return(nil)
				repo.EXPECT().
					SaveRefreshToken(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("redis error"))
			},
			wantErr: entity.ErrSaveRefreshToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, mockClient, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo, mockClient)

			resp, err := uc.RegisterUser(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.ID)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			}
		})
	}
}

// ====== LoginAccount ======

func TestLoginAccount(t *testing.T) {
	tests := []struct {
		name      string
		login     string
		password  string
		mockSetup func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider)
		wantErr   error
	}{
		{
			name:     "success",
			login:    "test1",
			password: "Qwerty12",
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				client.EXPECT().
					CompareAuthData(gomock.Any(), gomock.Any()).
					Return(&usersclient.AuthResponse{
						ID:       "some-uuid",
						Username: "TestUser",
					}, nil)
				repo.EXPECT().
					SaveRefreshToken(gomock.Any(), "some-uuid", gomock.Any()).
					Return(nil)
				client.EXPECT().
					UpdateRefreshToken(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:     "invalid credentials",
			login:    "test1",
			password: "wrongpass",
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				client.EXPECT().
					CompareAuthData(gomock.Any(), gomock.Any()).
					Return(nil, entity.ErrorResponse{
						ErrorDetail: entity.ErrorDetail{
							Code:    "INVALID_CREDENTIALS",
							Message: "Invalid login or password",
						},
					})
			},
			wantErr: entity.ErrorResponse{},
		},
		{
			name:     "save refresh token failed",
			login:    "test1",
			password: "Qwerty12",
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				client.EXPECT().
					CompareAuthData(gomock.Any(), gomock.Any()).
					Return(&usersclient.AuthResponse{
						ID:       "some-uuid",
						Username: "TestUser",
					}, nil)
				repo.EXPECT().
					SaveRefreshToken(gomock.Any(), "some-uuid", gomock.Any()).
					Return(errors.New("redis error"))
			},
			wantErr: entity.ErrUpdateRefreshToken,
		},
		{
			name:     "update refresh token in client failed",
			login:    "test1",
			password: "Qwerty12",
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				client.EXPECT().
					CompareAuthData(gomock.Any(), gomock.Any()).
					Return(&usersclient.AuthResponse{
						ID:       "some-uuid",
						Username: "TestUser",
					}, nil)
				repo.EXPECT().
					SaveRefreshToken(gomock.Any(), "some-uuid", gomock.Any()).
					Return(nil)
				client.EXPECT().
					UpdateRefreshToken(gomock.Any(), gomock.Any()).
					Return(errors.New("client error"))
			},
			wantErr: entity.ErrUpdateRefreshToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, mockClient, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo, mockClient)

			resp, err := uc.LoginAccount(context.Background(), tt.login, tt.password)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			}
		})
	}
}

// ====== RefreshTokens ======

func TestRefreshTokens(t *testing.T) {
	validToken := generateValidToken("some-uuid")

	tests := []struct {
		name         string
		refreshToken string
		authHeader   string
		mockSetup    func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider)
		wantErr      error
	}{
		{
			name:         "success",
			refreshToken: "stored-refresh-token",
			authHeader:   "Bearer " + validToken,
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				repo.EXPECT().
					GetRefreshToken(gomock.Any(), "some-uuid").
					Return("stored-refresh-token", nil)
				repo.EXPECT().
					SaveRefreshToken(gomock.Any(), "some-uuid", gomock.Any()).
					Return(nil)
				client.EXPECT().
					UpdateRefreshToken(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:         "invalid auth header",
			refreshToken: "stored-refresh-token",
			authHeader:   "InvalidHeader",
			mockSetup:    func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {},
			wantErr:      entity.ErrInvalidAuthHeader,
		},
		{
			name:         "invalid access token",
			refreshToken: "stored-refresh-token",
			authHeader:   "Bearer invalid-token",
			mockSetup:    func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {},
			wantErr:      entity.ErrInvalidToken,
		},
		{
			name:         "refresh token mismatch",
			refreshToken: "wrong-refresh-token",
			authHeader:   "Bearer " + validToken,
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				repo.EXPECT().
					GetRefreshToken(gomock.Any(), "some-uuid").
					Return("stored-refresh-token", nil)
			},
			wantErr: entity.ErrRefreshTokenMismatch,
		},
		{
			name:         "save new refresh token failed",
			refreshToken: "stored-refresh-token",
			authHeader:   "Bearer " + validToken,
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				repo.EXPECT().
					GetRefreshToken(gomock.Any(), "some-uuid").
					Return("stored-refresh-token", nil)
				repo.EXPECT().
					SaveRefreshToken(gomock.Any(), "some-uuid", gomock.Any()).
					Return(errors.New("redis error"))
			},
			wantErr: entity.ErrSaveRefreshToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, mockClient, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo, mockClient)

			resp, err := uc.RefreshTokens(context.Background(), tt.refreshToken, tt.authHeader)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			}
		})
	}
}
