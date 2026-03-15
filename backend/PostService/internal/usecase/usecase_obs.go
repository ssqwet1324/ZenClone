package usecase

import (
	"PostService/internal/entity"
	pkgMetrics "PostService/internal/pkg/prometheus"
	"context"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const nameTracer = "usecase_post"

// PostUseCaseObs структура обертки
type PostUseCaseObs struct {
	*PostUseCase
	metrics *pkgMetrics.Metrics
}

// NewObs  конструктор
func NewObs(uc *PostUseCase) *PostUseCaseObs {
	return &PostUseCaseObs{
		PostUseCase: uc,
		metrics:     pkgMetrics.NewMetrics(nameTracer),
	}
}

// observe сбор данных
func (uc *PostUseCaseObs) observe(ctx context.Context, method string, fn func(context.Context) error) error {
	ctx, span := otel.Tracer(nameTracer).Start(ctx, method)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start).Seconds()

	if err != nil {
		uc.metrics.HitError(method)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to uc.PostUseCase."+method)
	} else {
		uc.metrics.HitSuccess(method)
	}

	uc.metrics.HitDuration(method, duration)

	return err
}

// CreatePost - метрики + трассировка
func (uc *PostUseCaseObs) CreatePost(ctx context.Context, creatorUserID uuid.UUID, createPost entity.CreatePostRequest) (*entity.CreatePostResponse, error) {
	const method = "CreatePost"
	var resp *entity.CreatePostResponse

	err := uc.observe(ctx, method, func(ctx context.Context) error {
		r, err := uc.PostUseCase.CreatePost(ctx, creatorUserID, createPost)
		resp = r
		return err
	})

	return resp, err
}

// UpdatePost - метрики + трассировка
func (uc *PostUseCaseObs) UpdatePost(ctx context.Context, postID uuid.UUID, authorID uuid.UUID, updateReq entity.UpdateUserPostRequest) error {
	const method = "UpdatePost"

	return uc.observe(ctx, method, func(ctx context.Context) error {
		return uc.PostUseCase.UpdatePost(ctx, postID, authorID, updateReq)
	})
}

// DeletePost - метрики + трассировка
func (uc *PostUseCaseObs) DeletePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error {
	const method = "DeletePost"

	return uc.observe(ctx, method, func(ctx context.Context) error {
		return uc.PostUseCase.DeletePost(ctx, postID, userID)
	})
}

// GetPostsUser - метрики + трассировка
func (uc *PostUseCaseObs) GetPostsUser(ctx context.Context, authorID uuid.UUID, limit int, cursor *entity.PostCursor) (*entity.PostListResponse, error) {
	const method = "GetPostsUser"
	var resp *entity.PostListResponse

	err := uc.observe(ctx, method, func(ctx context.Context) error {
		r, err := uc.PostUseCase.GetPostsUser(ctx, authorID, limit, cursor)
		resp = r
		return err
	})

	return resp, err
}

// GetPostsFromFeed - метрики + трассировка
func (uc *PostUseCaseObs) GetPostsFromFeed(ctx context.Context, userUsername string, userID uuid.UUID, accessToken string, limit int, cursor *entity.PostCursor) (*entity.PostListResponseFromFeed, error) {
	const method = "GetPostsFromFeed"
	var resp *entity.PostListResponseFromFeed

	err := uc.observe(ctx, method, func(ctx context.Context) error {
		r, err := uc.PostUseCase.GetPostsFromFeed(ctx, userUsername, userID, accessToken, limit, cursor)
		resp = r
		return err
	})

	return resp, err
}
