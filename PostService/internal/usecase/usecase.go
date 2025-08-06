package usecase

import (
	"PostService/internal/config"
	"PostService/internal/entity"
	"PostService/internal/kafka"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
)

type RepositoryProvider interface {
	CreatePost(ctx context.Context, createPost entity.CreatePostRequest) (*entity.CreatePostResponse, error)
	UpdatePost(ctx context.Context, postID uuid.UUID, authorID uuid.UUID, updateReq entity.UpdateUserPostRequest) (*entity.UpdateUserPostResponse, error)
	DeletePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error
	GetPostsUser(ctx context.Context, authorID uuid.UUID) (*entity.PostListResponse, error)
}

type PostUseCase struct {
	producer *kafka.Producer
	repo     RepositoryProvider
	cfg      *config.Config
}

func New(repo RepositoryProvider, cfg *config.Config, producer *kafka.Producer) *PostUseCase {
	return &PostUseCase{
		producer: producer,
		repo:     repo,
		cfg:      cfg,
	}
}

// CreatePost - создаем пост и отправляем в FeedService
func (s *PostUseCase) CreatePost(ctx context.Context, createPost entity.CreatePostRequest) (uuid.UUID, error) {
	postID := uuid.New()
	createPost.ID = postID

	data, err := s.repo.CreatePost(ctx, createPost)
	if err != nil {
		return uuid.Nil, fmt.Errorf("CreatePost usecase: error create data: %w", err)
	}

	postBytes, err := json.Marshal(data)
	if err != nil {
		return uuid.Nil, fmt.Errorf("CreatePost usecase: error marshal data: %w", err)
	}

	err = s.producer.WriteMessages(ctx, postID.String(), postBytes)
	if err != nil {
		return uuid.Nil, fmt.Errorf("CreatePost usecase: error write data: %w", err)
	}

	return createPost.ID, nil
}

// UpdatePost - обновляем пост и отправляем в FeedService
func (s *PostUseCase) UpdatePost(ctx context.Context, postID uuid.UUID, authorID uuid.UUID, updateReq entity.UpdateUserPostRequest) error {
	data, err := s.repo.UpdatePost(ctx, postID, authorID, updateReq)
	if err != nil {
		return fmt.Errorf("UpdatePost usecase: error update data: %w", err)
	}

	postBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("UpdatePost usecase: error marshal data: %w", err)
	}

	err = s.producer.WriteMessages(ctx, postID.String(), postBytes)
	if err != nil {
		return fmt.Errorf("UpdatePost usecase: error write data: %w", err)
	}

	return nil
}

// DeletePost - удаляем пост и отправляем его id в FeedService
func (s *PostUseCase) DeletePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error {
	err := s.repo.DeletePost(ctx, postID, userID)
	if err != nil {
		return fmt.Errorf("DeletePost usecase: error delete post: %w", err)
	}

	err = s.producer.WriteMessages(ctx, postID.String(), []byte("deleted post:"+postID.String()))
	if err != nil {
		return fmt.Errorf("DeletePost usecase: error write data: %w", err)
	}

	return nil
}

func (s *PostUseCase) GetPostsUser(ctx context.Context, authorID uuid.UUID) (*entity.PostListResponse, error) {
	rows, err := s.repo.GetPostsUser(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("GetPostsUser usecase: error get posts user: %w", err)
	}

	if rows.Posts == nil {
		return &entity.PostListResponse{}, fmt.Errorf("GetPostsUser usecase: posts user not found")
	}

	return rows, nil
}
