package usecase

import (
	"PostService/internal/config"
	"PostService/internal/entity"
	"PostService/internal/kafka"
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RepositoryProvider - функции из repository
type RepositoryProvider interface {
	CreatePost(ctx context.Context, createPost entity.CreatePostResponse) (*entity.CreatePostResponse, error)
	UpdatePost(ctx context.Context, postID uuid.UUID, authorID uuid.UUID, updateReq entity.UpdateUserPostRequest) (*entity.UpdateUserPostResponse, error)
	DeletePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error
	GetPostsUser(ctx context.Context, authorID uuid.UUID) (*entity.PostListResponse, error)
}

// PostUseCase - структура бизнес логики
type PostUseCase struct {
	producer *kafka.Producer
	repo     RepositoryProvider
	cfg      *config.Config
	log      *zap.Logger
}

// New - конструктор бизнес логики
func New(repo RepositoryProvider, cfg *config.Config, producer *kafka.Producer, log *zap.Logger) *PostUseCase {
	return &PostUseCase{
		producer: producer,
		repo:     repo,
		cfg:      cfg,
		log:      log.Named("UseCase"),
	}
}

// CreatePost - создаем пост и отправляем в FeedService
func (s *PostUseCase) CreatePost(ctx context.Context, creatorUserID uuid.UUID, createPost entity.CreatePostRequest) (*entity.CreatePostResponse, error) {
	// Валидация данных
	if createPost.Title == "" {
		s.log.Error("CreatePost: empty title", zap.String("author_id", creatorUserID.String()))
		return nil, entity.ErrEmptyTitle
	}

	if createPost.Content == "" {
		s.log.Error("CreatePost: empty content", zap.String("author_id", creatorUserID.String()))
		return nil, entity.ErrEmptyContent
	}

	var createPostResponse entity.CreatePostResponse

	postID := uuid.New()

	createPostResponse.ID = postID
	createPostResponse.Title = createPost.Title
	createPostResponse.Content = createPost.Content
	createPostResponse.AuthorID = creatorUserID

	createdPost, err := s.repo.CreatePost(ctx, createPostResponse)
	if err != nil {
		s.log.Error("CreatePost: error creating post in repository",
			zap.String("post_id", postID.String()),
			zap.String("author_id", creatorUserID.String()),
			zap.Error(err))
		return nil, entity.ErrInternalError
	}

	postBytes, err := json.Marshal(createdPost)
	if err != nil {
		s.log.Error("CreatePost: error marshaling post",
			zap.String("post_id", createdPost.ID.String()),
			zap.Error(err))
		return nil, entity.ErrInternalError
	}

	err = s.producer.WriteMessages(ctx, createdPost.ID.String(), postBytes)
	if err != nil {
		s.log.Error("CreatePost: error writing to kafka",
			zap.String("post_id", createdPost.ID.String()),
			zap.Error(err))
		return nil, entity.ErrInternalError
	}

	return createdPost, nil
}

// UpdatePost - обновляем пост и отправляем в FeedService
func (s *PostUseCase) UpdatePost(ctx context.Context, postID uuid.UUID, authorID uuid.UUID, updateReq entity.UpdateUserPostRequest) error {
	data, err := s.repo.UpdatePost(ctx, postID, authorID, updateReq)
	if err != nil {
		s.log.Error("UpdatePost: error updating post")
		return entity.ErrInternalError
	}

	postBytes, err := json.Marshal(data)
	if err != nil {
		s.log.Error("UpdatePost: error marshaling post",
			zap.String("post_id", postID.String()),
			zap.Error(err))
		return entity.ErrInternalError
	}

	err = s.producer.WriteMessages(ctx, postID.String(), postBytes)
	if err != nil {
		s.log.Error("UpdatePost: error writing to kafka",
			zap.String("post_id", postID.String()),
			zap.Error(err))
		return entity.ErrInternalError
	}

	return nil
}

// DeletePost - удаляем пост и отправляем его id в FeedService
func (s *PostUseCase) DeletePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error {
	err := s.repo.DeletePost(ctx, postID, userID)
	if err != nil {
		return entity.ErrInternalError
	}

	err = s.producer.WriteMessages(ctx, postID.String(), []byte("deleted post:"+postID.String()))
	if err != nil {
		s.log.Error("DeletePost: error writing to kafka",
			zap.String("post_id", postID.String()),
			zap.Error(err))
		return entity.ErrInternalError
	}

	return nil
}

// GetPostsUser - получаем посты пользователя
func (s *PostUseCase) GetPostsUser(ctx context.Context, authorID uuid.UUID) (*entity.PostListResponse, error) {
	rows, err := s.repo.GetPostsUser(ctx, authorID)
	if err != nil {
		s.log.Error("GetPostsUser: error getting posts from repository",
			zap.String("author_id", authorID.String()),
			zap.Error(err))
		return nil, entity.ErrInternalError
	}

	if len(rows.Posts) == 0 {
		s.log.Warn("GetPostsUser: posts not found",
			zap.String("author_id", authorID.String()))
		return &entity.PostListResponse{}, entity.ErrPostsNotFound
	}

	return rows, nil
}
