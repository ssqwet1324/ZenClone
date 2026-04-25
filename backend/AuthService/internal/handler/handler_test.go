package handler

import (
	"AuthService/internal/config"
	"AuthService/internal/entity"
	"AuthService/mocks"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func validRegisterBody() entity.RegisterRequest {
	return entity.RegisterRequest{
		Login:     "test1",
		Password:  "Qwerty12",
		Username:  "test",
		FirstName: "Test",
		LastName:  "Testov",
		Bio:       "I like tests",
	}
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		mockSetup      func(m *mocks.MockUseCaseInterface)
		expectedStatus int
		expectedCode   string
	}{
		{
			name: "success",
			body: validRegisterBody(),
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					RegisterUser(gomock.Any(), gomock.Any()).
					Return(&entity.RegisterResponse{
						ID:           "some-uuid",
						AccessToken:  "access-token",
						RefreshToken: "refresh-token",
						Username:     "test",
					}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid json",
			body: "not a json {{{{",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REQUEST",
		},
		{
			name: "validation failed",
			body: entity.RegisterRequest{
				Login:     "",
				Password:  "Qwerty12",
				Username:  "test",
				FirstName: "Test",
				LastName:  "Testov",
				Bio:       "I like tests",
			},
			mockSetup: func(m *mocks.MockUseCaseInterface) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "user already exists",
			body: validRegisterBody(),
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					RegisterUser(gomock.Any(), gomock.Any()).
					Return(nil, entity.ErrorResponse{
						ErrorDetail: entity.ErrorDetail{
							Code:    "USER_ALREADY_EXISTS",
							Message: "User already exists",
						},
					})
			},
			expectedStatus: http.StatusConflict,
			expectedCode:   "USER_ALREADY_EXISTS",
		},
		{
			name: "internal error - hash password",
			body: validRegisterBody(),
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					RegisterUser(gomock.Any(), gomock.Any()).
					Return(nil, entity.ErrHashPassword)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "PASSWORD_HASH_ERROR",
		},
		{
			name: "register user failed",
			body: validRegisterBody(),
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					RegisterUser(gomock.Any(), gomock.Any()).
					Return(nil, entity.ErrRegisterUser)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "REGISTRATION_FAILED",
		},
		{
			name: "generate access token failed",
			body: validRegisterBody(),
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					RegisterUser(gomock.Any(), gomock.Any()).
					Return(nil, entity.ErrGenerateAccessToken)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "TOKEN_GENERATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockUseCaseInterface(ctrl)
			tt.mockSetup(mockUC)

			h := New(mockUC, zap.NewNop(), &config.Config{}, nil)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/register", h.Register)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		mockSetup      func(m *mocks.MockUseCaseInterface)
		expectedStatus int
	}{
		{
			name: "success",
			body: entity.LoginUserInfo{
				Login:    "test1",
				Password: "Qwerty12",
			},
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					LoginAccount(gomock.Any(), "test1", "Qwerty12").
					Return(&entity.LoginResponse{
						ID:           "some-uuid",
						AccessToken:  "access-token",
						RefreshToken: "refresh-token",
						Username:     "test",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid json",
			body: "not a json {{{{",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "validation failed - empty login",
			body: entity.LoginUserInfo{
				Login:    "",
				Password: "Qwerty12",
			},
			mockSetup: func(m *mocks.MockUseCaseInterface) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			body: entity.LoginUserInfo{Login: "test1", Password: "Qwerty12"},
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					LoginAccount(gomock.Any(), "test1", "Qwerty12").
					Return(nil, entity.ErrorResponse{
						ErrorDetail: entity.ErrorDetail{
							Code:    "INVALID_CREDENTIALS",
							Message: "Invalid login or password",
						},
					})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user not found",
			body: entity.LoginUserInfo{Login: "test1", Password: "Qwerty12"},
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					LoginAccount(gomock.Any(), "test1", "Qwerty12").
					Return(nil, entity.ErrorResponse{
						ErrorDetail: entity.ErrorDetail{
							Code:    "USER_NOT_FOUND",
							Message: "User not found",
						},
					})
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "compare auth data failed",
			body: entity.LoginUserInfo{Login: "test1", Password: "Qwerty12"},
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					LoginAccount(gomock.Any(), "test1", "Qwerty12").
					Return(nil, entity.ErrCompareAuthData)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "internal error",
			body: entity.LoginUserInfo{Login: "test1", Password: "Qwerty12"},
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					LoginAccount(gomock.Any(), "test1", "Qwerty12").
					Return(nil, entity.ErrUpdateRefreshToken)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockUseCaseInterface(ctrl)
			tt.mockSetup(mockUC)

			h := New(mockUC, zap.NewNop(), &config.Config{}, nil)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/login", h.Login)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRefresh(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		authHeader     string
		mockSetup      func(m *mocks.MockUseCaseInterface)
		expectedStatus int
	}{
		{
			name: "success",
			body: entity.TokenRequest{
				RefreshToken: "valid-refresh-token",
			},
			authHeader: "Bearer valid-access-token",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					RefreshTokens(gomock.Any(), "valid-refresh-token", "Bearer valid-access-token").
					Return(&entity.RefreshResponse{
						AccessToken:  "new-access-token",
						RefreshToken: "new-refresh-token",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "invalid json",
			body:       "not a json {{{{",
			authHeader: "Bearer valid-access-token",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing refresh token",
			body: entity.TokenRequest{
				RefreshToken: "",
			},
			authHeader: "Bearer valid-access-token",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing auth header",
			body: entity.TokenRequest{
				RefreshToken: "valid-refresh-token",
			},
			authHeader: "",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid token",
			body:       entity.TokenRequest{RefreshToken: "valid-refresh-token"},
			authHeader: "Bearer invalid-access-token",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					RefreshTokens(gomock.Any(), "valid-refresh-token", "Bearer invalid-access-token").
					Return(nil, entity.ErrInvalidToken)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "refresh token mismatch",
			body:       entity.TokenRequest{RefreshToken: "wrong-refresh-token"},
			authHeader: "Bearer valid-access-token",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					RefreshTokens(gomock.Any(), "wrong-refresh-token", "Bearer valid-access-token").
					Return(nil, entity.ErrRefreshTokenMismatch)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "get refresh token failed",
			body:       entity.TokenRequest{RefreshToken: "some-token"},
			authHeader: "Bearer valid-token",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					RefreshTokens(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, entity.ErrGetRefreshToken)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "default internal error",
			body:       entity.TokenRequest{RefreshToken: "some-token"},
			authHeader: "Bearer valid-token",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					RefreshTokens(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("unknown error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockUseCaseInterface(ctrl)
			tt.mockSetup(mockUC)

			h := New(mockUC, zap.NewNop(), &config.Config{}, nil)

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.POST("/refresh", h.Refresh)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
