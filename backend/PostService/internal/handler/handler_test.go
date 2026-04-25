package handler

import (
	"PostService/internal/entity"
	"PostService/mocks"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func setupRouter(h *PostHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func withUserID(userID string) func(*gin.Context) {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}
}

// ====== CreatePost ======

func TestCreatePost(t *testing.T) {
	validUserID := uuid.New().String()

	tests := []struct {
		name           string
		userID         string
		body           any
		mockSetup      func(m *mocks.MockUseCaseInterface)
		expectedStatus int
	}{
		{
			name:   "success",
			userID: validUserID,
			body:   entity.CreatePostRequest{Title: "Hello", Content: "World"},
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					CreatePost(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&entity.CreatePostResponse{
						ID:        uuid.New(),
						Title:     "Hello",
						Content:   "World",
						AuthorID:  uuid.New(),
						CreatedAt: time.Now(),
					}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "no userID in context",
			userID:         "",
			body:           entity.CreatePostRequest{Title: "Hello", Content: "World"},
			mockSetup:      func(m *mocks.MockUseCaseInterface) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid json",
			userID:         validUserID,
			body:           "not a json {{{",
			mockSetup:      func(m *mocks.MockUseCaseInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "empty title",
			userID: validUserID,
			body:   entity.CreatePostRequest{Title: "", Content: "World"},
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					CreatePost(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, entity.ErrEmptyTitle)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "empty content",
			userID: validUserID,
			body:   entity.CreatePostRequest{Title: "Hello", Content: ""},
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					CreatePost(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, entity.ErrEmptyContent)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "internal error",
			userID: validUserID,
			body:   entity.CreatePostRequest{Title: "Hello", Content: "World"},
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					CreatePost(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, entity.ErrInternalError)
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

			h := New(mockUC, zap.NewNop())
			router := setupRouter(h)

			if tt.userID != "" {
				router.POST("/posts/create", withUserID(tt.userID), h.CreatePost)
			} else {
				router.POST("/posts/create", h.CreatePost)
			}

			var bodyBytes []byte
			switch v := tt.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/posts/create", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ====== UpdatePost ======

func TestUpdatePost(t *testing.T) {
	validUserID := uuid.New().String()
	validPostID := uuid.New().String()
	title := "Updated Title"

	tests := []struct {
		name           string
		postID         string
		userID         string
		body           any
		mockSetup      func(m *mocks.MockUseCaseInterface)
		expectedStatus int
	}{
		{
			name:   "success",
			postID: validPostID,
			userID: validUserID,
			body:   entity.UpdateUserPostRequest{Title: &title},
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().UpdatePost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid postID",
			postID:         "not-a-uuid",
			userID:         validUserID,
			body:           entity.UpdateUserPostRequest{Title: &title},
			mockSetup:      func(m *mocks.MockUseCaseInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "no userID in context",
			postID:         validPostID,
			userID:         "",
			body:           entity.UpdateUserPostRequest{Title: &title},
			mockSetup:      func(m *mocks.MockUseCaseInterface) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "post not owned",
			postID: validPostID,
			userID: validUserID,
			body:   entity.UpdateUserPostRequest{Title: &title},
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().UpdatePost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(entity.ErrPostNotOwned)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "internal error",
			postID: validPostID,
			userID: validUserID,
			body:   entity.UpdateUserPostRequest{Title: &title},
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().UpdatePost(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(entity.ErrInternalError)
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

			h := New(mockUC, zap.NewNop())
			router := setupRouter(h)

			if tt.userID != "" {
				router.POST("/posts/update/:postID", withUserID(tt.userID), h.UpdatePost)
			} else {
				router.POST("/posts/update/:postID", h.UpdatePost)
			}

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/posts/update/"+tt.postID, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ====== DeletePost ======

func TestDeletePost(t *testing.T) {
	validUserID := uuid.New().String()
	validPostID := uuid.New().String()

	tests := []struct {
		name           string
		postID         string
		userID         string
		mockSetup      func(m *mocks.MockUseCaseInterface)
		expectedStatus int
	}{
		{
			name:   "success",
			postID: validPostID,
			userID: validUserID,
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().DeletePost(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid postID",
			postID:         "not-a-uuid",
			userID:         validUserID,
			mockSetup:      func(m *mocks.MockUseCaseInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "no userID in context",
			postID:         validPostID,
			userID:         "",
			mockSetup:      func(m *mocks.MockUseCaseInterface) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "post not owned",
			postID: validPostID,
			userID: validUserID,
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().DeletePost(gomock.Any(), gomock.Any(), gomock.Any()).Return(entity.ErrPostNotOwned)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "internal error",
			postID: validPostID,
			userID: validUserID,
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().DeletePost(gomock.Any(), gomock.Any(), gomock.Any()).Return(entity.ErrInternalError)
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

			h := New(mockUC, zap.NewNop())
			router := setupRouter(h)

			if tt.userID != "" {
				router.DELETE("/posts/delete/:postID", withUserID(tt.userID), h.DeletePost)
			} else {
				router.DELETE("/posts/delete/:postID", h.DeletePost)
			}

			req := httptest.NewRequest(http.MethodDelete, "/posts/delete/"+tt.postID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ====== GetPostsUser ======

func TestGetPostsUser(t *testing.T) {
	validUserID := uuid.New().String()

	tests := []struct {
		name           string
		userID         string
		query          string
		mockSetup      func(m *mocks.MockUseCaseInterface)
		expectedStatus int
	}{
		{
			name:   "success",
			userID: validUserID,
			query:  "",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					GetPostsUser(gomock.Any(), gomock.Any(), 20, gomock.Any()).
					Return(&entity.PostListResponse{
						Posts: []entity.PostResponse{{ID: uuid.New(), Title: "Post 1"}},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid userID",
			userID:         "not-a-uuid",
			query:          "",
			mockSetup:      func(m *mocks.MockUseCaseInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid limit",
			userID:         validUserID,
			query:          "?limit=abc",
			mockSetup:      func(m *mocks.MockUseCaseInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "posts not found",
			userID: validUserID,
			query:  "",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					GetPostsUser(gomock.Any(), gomock.Any(), 20, gomock.Any()).
					Return(nil, entity.ErrPostsNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "internal error",
			userID: validUserID,
			query:  "",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					GetPostsUser(gomock.Any(), gomock.Any(), 20, gomock.Any()).
					Return(nil, entity.ErrInternalError)
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

			h := New(mockUC, zap.NewNop())
			router := setupRouter(h)
			router.GET("/posts/by-user/:userID", h.GetPostsUser)

			req := httptest.NewRequest(http.MethodGet, "/posts/by-user/"+tt.userID+tt.query, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// ====== GetPostsFeedFromUser ======

func TestGetPostsFeedFromUser(t *testing.T) {
	validUserID := uuid.New().String()

	tests := []struct {
		name           string
		userID         string
		authHeader     string
		query          string
		mockSetup      func(m *mocks.MockUseCaseInterface)
		expectedStatus int
	}{
		{
			name:       "success",
			userID:     validUserID,
			authHeader: "Bearer valid-token",
			query:      "?username=testuser",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					GetPostsFromFeed(gomock.Any(), "testuser", gomock.Any(), "valid-token", 20, gomock.Any()).
					Return(&entity.PostListResponseFromFeed{
						Posts: []entity.PostResponseFromFeed{{ID: uuid.New(), Title: "Feed Post"}},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no userID in context",
			userID:         "",
			authHeader:     "Bearer valid-token",
			query:          "?username=testuser",
			mockSetup:      func(m *mocks.MockUseCaseInterface) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing auth header",
			userID:         validUserID,
			authHeader:     "",
			query:          "?username=testuser",
			mockSetup:      func(m *mocks.MockUseCaseInterface) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing username query param",
			userID:         validUserID,
			authHeader:     "Bearer valid-token",
			query:          "",
			mockSetup:      func(m *mocks.MockUseCaseInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "posts not found",
			userID:     validUserID,
			authHeader: "Bearer valid-token",
			query:      "?username=testuser",
			mockSetup: func(m *mocks.MockUseCaseInterface) {
				m.EXPECT().
					GetPostsFromFeed(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, entity.ErrPostsNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockUseCaseInterface(ctrl)
			tt.mockSetup(mockUC)

			h := New(mockUC, zap.NewNop())
			router := setupRouter(h)

			if tt.userID != "" {
				router.GET("/posts/feed", withUserID(tt.userID), h.GetPostsFeedFromUser)
			} else {
				router.GET("/posts/feed", h.GetPostsFeedFromUser)
			}

			req := httptest.NewRequest(http.MethodGet, "/posts/feed"+tt.query, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
