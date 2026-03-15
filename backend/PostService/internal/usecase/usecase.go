package usecase

import (
	"PostService/internal/client/subsclient"
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
	GetPostsUser(ctx context.Context, authorID uuid.UUID, limit int, cursor *entity.PostCursor) (*entity.PostListResponse, error)
	GetPostsFromFeed(ctx context.Context, authorIDs []uuid.UUID, limit int, cursor *entity.PostCursor) (*entity.PostListResponseFromFeed, error)
}

type UseCaseInterface interface {
	CreatePost(ctx context.Context, creatorUserID uuid.UUID, createPost entity.CreatePostRequest) (*entity.CreatePostResponse, error)
	UpdatePost(ctx context.Context, postID uuid.UUID, authorID uuid.UUID, updateReq entity.UpdateUserPostRequest) error
	DeletePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error
	GetPostsUser(ctx context.Context, authorID uuid.UUID, limit int, cursor *entity.PostCursor) (*entity.PostListResponse, error)
	GetPostsFromFeed(ctx context.Context, userUsername string, userID uuid.UUID, accessToken string, limit int, cursor *entity.PostCursor) (*entity.PostListResponseFromFeed, error)
}

// PostUseCase - структура бизнес логики
type PostUseCase struct {
	producer *kafka.Producer
	repo     RepositoryProvider
	cfg      *config.Config
	log      *zap.Logger
	client   subsclient.ClientProvider
}

// New - конструктор бизнес логики
func New(repo RepositoryProvider, cfg *config.Config, producer *kafka.Producer,
	log *zap.Logger, client subsclient.ClientProvider) UseCaseInterface {
	usecase := &PostUseCase{
		producer: producer,
		repo:     repo,
		cfg:      cfg,
		client:   client,
		log:      log.Named("UseCase"),
	}

	return NewObs(usecase)
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
		s.log.Error("UpdatePost: error updating post", zap.Error(err))
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
func (s *PostUseCase) GetPostsUser(ctx context.Context, authorID uuid.UUID, limit int, cursor *entity.PostCursor) (*entity.PostListResponse, error) {
	rows, err := s.repo.GetPostsUser(ctx, authorID, limit, cursor)
	if err != nil {
		s.log.Error("GetPostsUser: error getting posts from repository",
			zap.String("author_id", authorID.String()),
			zap.Error(err))
		return nil, entity.ErrInternalError
	}

	if len(rows.Posts) == 0 {
		s.log.Warn("GetPostsUser: posts not found",
			zap.String("author_id", authorID.String()))
		return nil, entity.ErrPostsNotFound
	}

	return rows, nil
}

// GetPostsFromFeed - получаем посты для ленты
func (s *PostUseCase) GetPostsFromFeed(ctx context.Context, userUsername string, userID uuid.UUID, accessToken string, limit int, cursor *entity.PostCursor) (*entity.PostListResponseFromFeed, error) {
	// получаем подписки
	userSubs, err := s.client.GetSubsUser(ctx, userUsername, accessToken)
	if err != nil {
		s.log.Error("GetPostsFromFeed: error getting user subs from users service",
			zap.String("username", userUsername), zap.Error(err))
		return nil, err
	}

	if userSubs == nil || len(userSubs.Subs) == 0 {
		s.log.Warn("GetPostsFromFeed: user has no subscriptions",
			zap.String("username", userUsername))
		return nil, entity.ErrPostsNotFound
	}

	// собираем id подписок
	authorIDs := make([]uuid.UUID, 0, len(userSubs.Subs))
	for _, sub := range userSubs.Subs {
		authorIDs = append(authorIDs, sub.ID)
	}

	// получаем посты
	subsPosts, err := s.repo.GetPostsFromFeed(ctx, authorIDs, limit, cursor)
	if err != nil {
		s.log.Error("GetPostsFromFeed: error getting posts from repository",
			zap.String("user_id", userID.String()), zap.Error(err))
		return nil, entity.ErrInternalError
	}

	if len(subsPosts.Posts) == 0 {
		s.log.Warn("GetPostsFromFeed: posts not found",
			zap.String("user_id", userID.String()))
		return nil, entity.ErrPostsNotFound
	}

	// Создаем маппинг authorID -> username из подписок
	authorIDToUsername := make(map[uuid.UUID]string)
	for _, sub := range userSubs.Subs {
		authorIDToUsername[sub.ID] = sub.Username
	}

	// Собираем уникальные username авторов для запроса профилей
	uniqueUsernames := make(map[string]bool)
	for _, post := range subsPosts.Posts {
		if username, exists := authorIDToUsername[post.AuthorID]; exists {
			uniqueUsernames[username] = true
		}
	}

	// Получаем профили авторов
	authorProfiles := make(map[string]string)
	for username := range uniqueUsernames {
		profile, err := s.client.GetUserProfileByUsername(ctx, username, accessToken)
		if err != nil {
			s.log.Warn("GetPostsFromFeed: failed to get author profile",
				zap.String("username", username),
				zap.Error(err))
			continue
		}

		if profile != nil {
			authorProfiles[username] = profile.UserAvatarURL
		}
	}

	// Обогащаем посты данными авторов (username и avatar)
	for i := range subsPosts.Posts {
		authorID := subsPosts.Posts[i].AuthorID
		if username, exists := authorIDToUsername[authorID]; exists {
			subsPosts.Posts[i].AuthorName = username
			if avatarURL, hasAvatar := authorProfiles[username]; hasAvatar {
				subsPosts.Posts[i].AuthorAvatar = avatarURL
			}
		}
	}

	return subsPosts, nil
}
