package usecase

import (
	"PostService/internal/client/subsclient"
	"PostService/internal/config"
	"PostService/internal/entity"
	"PostService/internal/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func newTestUseCase(t *testing.T) (*PostUseCase, *mocks.MockRepositoryProvider, *mocks.MockProducerProvider, *mocks.MockClientProvider, *gomock.Controller) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockRepositoryProvider(ctrl)
	mockProducer := mocks.NewMockProducerProvider(ctrl)
	mockClient := mocks.NewMockClientProvider(ctrl)

	uc := &PostUseCase{
		repo:     mockRepo,
		producer: mockProducer,
		client:   mockClient,
		cfg:      &config.Config{},
		log:      zap.NewNop(),
	}

	return uc, mockRepo, mockProducer, mockClient, ctrl
}

// ====== CreatePost ======

func TestCreatePost(t *testing.T) {
	authorID := uuid.New()

	tests := []struct {
		name      string
		req       entity.CreatePostRequest
		mockSetup func(repo *mocks.MockRepositoryProvider, producer *mocks.MockProducerProvider)
		wantErr   error
	}{
		{
			name: "success",
			req:  entity.CreatePostRequest{Title: "Hello", Content: "World"},
			mockSetup: func(repo *mocks.MockRepositoryProvider, producer *mocks.MockProducerProvider) {
				repo.EXPECT().
					CreatePost(gomock.Any(), gomock.Any()).
					Return(&entity.CreatePostResponse{
						ID:        uuid.New(),
						Title:     "Hello",
						Content:   "World",
						AuthorID:  authorID,
						CreatedAt: time.Now(),
					}, nil)
				producer.EXPECT().
					WriteMessages(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:      "empty title",
			req:       entity.CreatePostRequest{Title: "", Content: "World"},
			mockSetup: func(repo *mocks.MockRepositoryProvider, producer *mocks.MockProducerProvider) {},
			wantErr:   entity.ErrEmptyTitle,
		},
		{
			name:      "empty content",
			req:       entity.CreatePostRequest{Title: "Hello", Content: ""},
			mockSetup: func(repo *mocks.MockRepositoryProvider, producer *mocks.MockProducerProvider) {},
			wantErr:   entity.ErrEmptyContent,
		},
		{
			name: "repo error",
			req:  entity.CreatePostRequest{Title: "Hello", Content: "World"},
			mockSetup: func(repo *mocks.MockRepositoryProvider, producer *mocks.MockProducerProvider) {
				repo.EXPECT().
					CreatePost(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db error"))
			},
			wantErr: entity.ErrInternalError,
		},
		{
			name: "kafka error",
			req:  entity.CreatePostRequest{Title: "Hello", Content: "World"},
			mockSetup: func(repo *mocks.MockRepositoryProvider, producer *mocks.MockProducerProvider) {
				repo.EXPECT().
					CreatePost(gomock.Any(), gomock.Any()).
					Return(&entity.CreatePostResponse{
						ID:       uuid.New(),
						Title:    "Hello",
						Content:  "World",
						AuthorID: authorID,
					}, nil)
				producer.EXPECT().
					WriteMessages(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("kafka error"))
			},
			wantErr: entity.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, mockProducer, _, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo, mockProducer)

			resp, err := uc.CreatePost(context.Background(), authorID, tt.req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

// ====== UpdatePost ======

func TestUpdatePost(t *testing.T) {
	postID := uuid.New()
	authorID := uuid.New()
	title := "New Title"
	content := "New Content"

	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockRepositoryProvider, producer *mocks.MockProducerProvider)
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(repo *mocks.MockRepositoryProvider, producer *mocks.MockProducerProvider) {
				repo.EXPECT().
					UpdatePost(gomock.Any(), postID, authorID, gomock.Any()).
					Return(&entity.UpdateUserPostResponse{Title: &title, Content: &content}, nil)
				producer.EXPECT().
					WriteMessages(gomock.Any(), postID.String(), gomock.Any()).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "repo error",
			mockSetup: func(repo *mocks.MockRepositoryProvider, producer *mocks.MockProducerProvider) {
				repo.EXPECT().
					UpdatePost(gomock.Any(), postID, authorID, gomock.Any()).
					Return(nil, errors.New("db error"))
			},
			wantErr: entity.ErrInternalError,
		},
		{
			name: "kafka error",
			mockSetup: func(repo *mocks.MockRepositoryProvider, producer *mocks.MockProducerProvider) {
				repo.EXPECT().
					UpdatePost(gomock.Any(), postID, authorID, gomock.Any()).
					Return(&entity.UpdateUserPostResponse{Title: &title}, nil)
				producer.EXPECT().
					WriteMessages(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("kafka error"))
			},
			wantErr: entity.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, mockProducer, _, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo, mockProducer)

			err := uc.UpdatePost(context.Background(), postID, authorID, entity.UpdateUserPostRequest{Title: &title})

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ====== DeletePost ======

func TestDeletePost(t *testing.T) {
	postID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockRepositoryProvider, producer *mocks.MockProducerProvider)
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(repo *mocks.MockRepositoryProvider, producer *mocks.MockProducerProvider) {
				repo.EXPECT().DeletePost(gomock.Any(), postID, userID).Return(nil)
				producer.EXPECT().WriteMessages(gomock.Any(), postID.String(), gomock.Any()).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "repo error",
			mockSetup: func(repo *mocks.MockRepositoryProvider, producer *mocks.MockProducerProvider) {
				repo.EXPECT().DeletePost(gomock.Any(), postID, userID).Return(errors.New("db error"))
			},
			wantErr: entity.ErrInternalError,
		},
		{
			name: "kafka error",
			mockSetup: func(repo *mocks.MockRepositoryProvider, producer *mocks.MockProducerProvider) {
				repo.EXPECT().DeletePost(gomock.Any(), postID, userID).Return(nil)
				producer.EXPECT().WriteMessages(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("kafka error"))
			},
			wantErr: entity.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, mockProducer, _, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo, mockProducer)

			err := uc.DeletePost(context.Background(), postID, userID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ====== GetPostsUser ======

func TestGetPostsUser(t *testing.T) {
	authorID := uuid.New()

	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockRepositoryProvider)
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GetPostsUser(gomock.Any(), authorID, 20, gomock.Any()).
					Return(&entity.PostListResponse{
						Posts: []entity.PostResponse{{ID: uuid.New(), Title: "Post 1"}},
					}, nil)
			},
			wantErr: nil,
		},
		{
			name: "posts not found - empty list",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GetPostsUser(gomock.Any(), authorID, 20, gomock.Any()).
					Return(&entity.PostListResponse{Posts: []entity.PostResponse{}}, nil)
			},
			wantErr: entity.ErrPostsNotFound,
		},
		{
			name: "repo error",
			mockSetup: func(repo *mocks.MockRepositoryProvider) {
				repo.EXPECT().
					GetPostsUser(gomock.Any(), authorID, 20, gomock.Any()).
					Return(nil, errors.New("db error"))
			},
			wantErr: entity.ErrInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, _, _, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo)

			resp, err := uc.GetPostsUser(context.Background(), authorID, 20, nil)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

// ====== GetPostsFromFeed ======

func TestGetPostsFromFeed(t *testing.T) {
	userID := uuid.New()
	subID := uuid.New()
	token := "test-token"
	username := "testuser"

	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider)
		wantErr   error
	}{
		{
			name: "success",
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				client.EXPECT().
					GetSubsUser(gomock.Any(), username, token).
					Return(&subsclient.SubsList{
						Subs: []subsclient.SubUserInfo{{ID: subID, Username: "author1"}},
					}, nil)
				repo.EXPECT().
					GetPostsFromFeed(gomock.Any(), gomock.Any(), 20, gomock.Any()).
					Return(&entity.PostListResponseFromFeed{
						Posts: []entity.PostResponseFromFeed{{ID: uuid.New(), AuthorID: subID, Title: "Post"}},
					}, nil)
				client.EXPECT().
					GetUserProfileByUsername(gomock.Any(), "author1", token).
					Return(&subsclient.UserProfile{UserAvatarURL: "http://avatar.url"}, nil)
			},
			wantErr: nil,
		},
		{
			name: "client error getting subs",
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				client.EXPECT().
					GetSubsUser(gomock.Any(), username, token).
					Return(nil, errors.New("client error"))
			},
			wantErr: errors.New("client error"),
		},
		{
			name: "no subscriptions",
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				client.EXPECT().
					GetSubsUser(gomock.Any(), username, token).
					Return(&subsclient.SubsList{Subs: []subsclient.SubUserInfo{}}, nil)
			},
			wantErr: entity.ErrPostsNotFound,
		},
		{
			name: "repo error",
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				client.EXPECT().
					GetSubsUser(gomock.Any(), username, token).
					Return(&subsclient.SubsList{
						Subs: []subsclient.SubUserInfo{{ID: subID, Username: "author1"}},
					}, nil)
				repo.EXPECT().
					GetPostsFromFeed(gomock.Any(), gomock.Any(), 20, gomock.Any()).
					Return(nil, errors.New("db error"))
			},
			wantErr: entity.ErrInternalError,
		},
		{
			name: "empty posts in feed",
			mockSetup: func(repo *mocks.MockRepositoryProvider, client *mocks.MockClientProvider) {
				client.EXPECT().
					GetSubsUser(gomock.Any(), username, token).
					Return(&subsclient.SubsList{
						Subs: []subsclient.SubUserInfo{{ID: subID, Username: "author1"}},
					}, nil)
				repo.EXPECT().
					GetPostsFromFeed(gomock.Any(), gomock.Any(), 20, gomock.Any()).
					Return(&entity.PostListResponseFromFeed{Posts: []entity.PostResponseFromFeed{}}, nil)
			},
			wantErr: entity.ErrPostsNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, mockRepo, _, mockClient, ctrl := newTestUseCase(t)
			defer ctrl.Finish()

			tt.mockSetup(mockRepo, mockClient)

			resp, err := uc.GetPostsFromFeed(context.Background(), username, userID, token, 20, nil)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}
