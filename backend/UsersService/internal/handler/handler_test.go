package handler

import (
	"UsersService/internal/entity"
	"UsersService/mocks"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func withUserID(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}
}

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// ====== UpdateRefreshToken ======

func TestUpdateRefreshToken(t *testing.T) {
	validReq := entity.UpdateRefreshTokenRequest{ID: uuid.New(), RefreshToken: "some-token"}

	tests := []struct {
		name           string
		body           any
		mockSetup      func(m *mocks.MockUsecaseInterface)
		expectedStatus int
	}{
		{
			name: "success",
			body: validReq,
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().UpdateRefreshToken(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid json",
			body:           "not a json {{{",
			mockSetup:      func(m *mocks.MockUsecaseInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "usecase error",
			body: validReq,
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().UpdateRefreshToken(gomock.Any(), gomock.Any()).Return(entity.ErrFailedToUpdateRefreshToken)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockUsecaseInterface(ctrl)
			tt.mockSetup(mockUC)

			h := New(mockUC, zap.NewNop())
			router := newRouter()
			router.POST("/internal/update-refresh-token", h.UpdateRefreshToken)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/internal/update-refresh-token", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ====== GetRefreshToken ======

func TestGetRefreshToken(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name           string
		body           any
		mockSetup      func(m *mocks.MockUsecaseInterface)
		expectedStatus int
	}{
		{
			name: "success",
			body: entity.TokenRequest{ID: validID},
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().
					GetRefreshToken(gomock.Any(), validID).
					Return(&entity.TokenResponse{RefreshToken: "some-token"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid json",
			body:           "not json {{{",
			mockSetup:      func(m *mocks.MockUsecaseInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "usecase error",
			body: entity.TokenRequest{ID: validID},
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().
					GetRefreshToken(gomock.Any(), validID).
					Return(nil, entity.ErrFailedToGetRefreshToken)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockUsecaseInterface(ctrl)
			tt.mockSetup(mockUC)

			h := New(mockUC, zap.NewNop())
			router := newRouter()
			router.POST("/internal/get-refresh-token", h.GetRefreshToken)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/internal/get-refresh-token", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ====== CompareAuthPassword ======

func TestCompareAuthPassword(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		body           any
		mockSetup      func(m *mocks.MockUsecaseInterface)
		expectedStatus int
	}{
		{
			name: "success",
			body: entity.AuthRequest{Login: "test@test.com", Password: "pass"},
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().
					CompareAuthData(gomock.Any(), gomock.Any()).
					Return(&entity.CompareDataResponse{ID: userID, Username: "testuser"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid json",
			body:           "not json {{{",
			mockSetup:      func(m *mocks.MockUsecaseInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "user not found",
			body: entity.AuthRequest{Login: "notexist@test.com", Password: "pass"},
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().
					CompareAuthData(gomock.Any(), gomock.Any()).
					Return(nil, entity.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "incorrect password",
			body: entity.AuthRequest{Login: "test@test.com", Password: "wrongpass"},
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().
					CompareAuthData(gomock.Any(), gomock.Any()).
					Return(nil, entity.ErrIncorrectPassword)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockUsecaseInterface(ctrl)
			tt.mockSetup(mockUC)

			h := New(mockUC, zap.NewNop())
			router := newRouter()
			router.POST("/internal/compare-auth-data", h.CompareAuthPassword)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/internal/compare-auth-data", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ====== AddUser ======

func TestAddUser(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		mockSetup      func(m *mocks.MockUsecaseInterface)
		expectedStatus int
	}{
		{
			name: "success",
			body: entity.AddUserRequest{Username: "testuser", Login: "test@test.com", Password: "pass"},
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().AddUser(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid json",
			body:           "not json {{{",
			mockSetup:      func(m *mocks.MockUsecaseInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "user already exists",
			body: entity.AddUserRequest{Username: "existing", Login: "ex@test.com", Password: "pass"},
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().AddUser(gomock.Any(), gomock.Any()).Return(entity.ErrUserAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "internal error",
			body: entity.AddUserRequest{Username: "testuser", Login: "test@test.com", Password: "pass"},
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().AddUser(gomock.Any(), gomock.Any()).Return(entity.ErrInternalServer)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockUsecaseInterface(ctrl)
			tt.mockSetup(mockUC)

			h := New(mockUC, zap.NewNop())
			router := newRouter()
			router.POST("/internal/add-user", h.AddUser)

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/internal/add-user", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ====== GetProfile ======

func TestGetProfile(t *testing.T) {
	yourUserID := uuid.New().String()
	otherUserID := uuid.New().String()

	tests := []struct {
		name           string
		jwtUserID      string
		username       string
		mockSetup      func(m *mocks.MockUsecaseInterface)
		expectedStatus int
	}{
		{
			name:      "success",
			jwtUserID: yourUserID,
			username:  "testuser",
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().
					GetUserIDByUsername(gomock.Any(), "testuser").
					Return(otherUserID, nil)
				m.EXPECT().
					GetUserProfileByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&entity.ProfileUserInfoResponse{FirstName: "John"}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no jwt userID",
			jwtUserID:      "",
			username:       "testuser",
			mockSetup:      func(m *mocks.MockUsecaseInterface) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "user not found",
			jwtUserID: yourUserID,
			username:  "notexist",
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().
					GetUserIDByUsername(gomock.Any(), "notexist").
					Return("", entity.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockUsecaseInterface(ctrl)
			tt.mockSetup(mockUC)

			h := New(mockUC, zap.NewNop())
			router := newRouter()

			if tt.jwtUserID != "" {
				router.GET("/api/v1/get-user-profile/:username", withUserID(tt.jwtUserID), h.GetProfile)
			} else {
				router.GET("/api/v1/get-user-profile/:username", h.GetProfile)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/get-user-profile/"+tt.username, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ====== Subscribe ======

func TestSubscribe(t *testing.T) {
	followerID := uuid.New().String()
	followingID := uuid.New().String()

	tests := []struct {
		name           string
		jwtUserID      string
		username       string
		mockSetup      func(m *mocks.MockUsecaseInterface)
		expectedStatus int
	}{
		{
			name:      "success",
			jwtUserID: followerID,
			username:  "targetuser",
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().
					GetUserIDByUsername(gomock.Any(), "targetuser").
					Return(followingID, nil)
				m.EXPECT().
					SubscribeToUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no jwt userID",
			jwtUserID:      "",
			username:       "targetuser",
			mockSetup:      func(m *mocks.MockUsecaseInterface) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "target user not found",
			jwtUserID: followerID,
			username:  "notexist",
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().
					GetUserIDByUsername(gomock.Any(), "notexist").
					Return("", entity.ErrUserNotFound)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "subscribe error",
			jwtUserID: followerID,
			username:  "targetuser",
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().
					GetUserIDByUsername(gomock.Any(), "targetuser").
					Return(followingID, nil)
				m.EXPECT().
					SubscribeToUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(entity.ErrFailedToSubscribe)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockUsecaseInterface(ctrl)
			tt.mockSetup(mockUC)

			h := New(mockUC, zap.NewNop())
			router := newRouter()

			if tt.jwtUserID != "" {
				router.POST("/api/v1/user/subscribe/:username", withUserID(tt.jwtUserID), h.Subscribe)
			} else {
				router.POST("/api/v1/user/subscribe/:username", h.Subscribe)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/user/subscribe/"+tt.username, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ====== GetSubsUser ======

func TestGetSubsUser(t *testing.T) {
	jwtUserID := uuid.New().String()
	targetUserID := uuid.New().String()

	tests := []struct {
		name           string
		jwtUserID      string
		username       string
		mockSetup      func(m *mocks.MockUsecaseInterface)
		expectedStatus int
	}{
		{
			name:      "success",
			jwtUserID: jwtUserID,
			username:  "testuser",
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().
					GetUserIDByUsername(gomock.Any(), "testuser").
					Return(targetUserID, nil)
				m.EXPECT().
					GetSubsUser(gomock.Any(), gomock.Any()).
					Return(&entity.SubsList{Subs: []entity.SubUserInfo{{Username: "sub1"}}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no jwt userID",
			jwtUserID:      "",
			username:       "testuser",
			mockSetup:      func(m *mocks.MockUsecaseInterface) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "no subscriptions",
			jwtUserID: jwtUserID,
			username:  "testuser",
			mockSetup: func(m *mocks.MockUsecaseInterface) {
				m.EXPECT().
					GetUserIDByUsername(gomock.Any(), "testuser").
					Return(targetUserID, nil)
				m.EXPECT().
					GetSubsUser(gomock.Any(), gomock.Any()).
					Return(nil, entity.ErrNoSubscriptions)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockUsecaseInterface(ctrl)
			tt.mockSetup(mockUC)

			h := New(mockUC, zap.NewNop())
			router := newRouter()

			if tt.jwtUserID != "" {
				router.GET("/api/v1/user/subs/:username", withUserID(tt.jwtUserID), h.GetSubsUser)
			} else {
				router.GET("/api/v1/user/subs/:username", h.GetSubsUser)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/user/subs/"+tt.username, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
