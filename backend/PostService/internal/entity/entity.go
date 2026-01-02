package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreatePostResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorID  uuid.UUID `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateUserPostRequest struct {
	Title     *string   `json:"title"`
	Content   *string   `json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateUserPostResponse struct {
	Title     *string   `json:"title"`
	Content   *string   `json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostListResponse struct {
	Posts      []PostResponse `json:"posts"`
	NextCursor *PostCursor    `json:"next_cursor"`
}

// PostCursor используется для бесконечной ленты постов в профиле
type PostCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

// CreatePostResponseData - данные ответа при создании поста
type CreatePostResponseData struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	AuthorID  string `json:"author_id"`
	CreatedAt string `json:"created_at"`
}

// CreatePostSuccessResponse - успешный ответ при создании поста
type CreatePostSuccessResponse struct {
	Message string                 `json:"message"`
	Data    CreatePostResponseData `json:"data"`
}

// UpdatePostResponseData - данные ответа при обновлении поста
type UpdatePostResponseData struct {
	PostID string `json:"post_id"`
}

// UpdatePostSuccessResponse - успешный ответ при обновлении поста
type UpdatePostSuccessResponse struct {
	Message string                 `json:"message"`
	Data    UpdatePostResponseData `json:"data"`
}

// DeletePostResponseData - данные ответа при удалении поста
type DeletePostResponseData struct {
	PostID string `json:"post_id"`
}

// DeletePostSuccessResponse - успешный ответ при удалении поста
type DeletePostSuccessResponse struct {
	Message string                 `json:"message"`
	Data    DeletePostResponseData `json:"data"`
}

// GetPostsUserResponseData - данные ответа при получении постов
type GetPostsUserResponseData struct {
	Posts      []PostResponse `json:"posts"`
	Count      int            `json:"count"`
	NextCursor string         `json:"next_cursor"`
}

// GetPostsUserSuccessResponse - успешный ответ при получении постов
type GetPostsUserSuccessResponse struct {
	Message string                   `json:"message"`
	Data    GetPostsUserResponseData `json:"data"`
}

// ErrorResponse - ответ ошибки
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail - информация об ошибке
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

var (
	// ErrPostNotOwned - пост не принадлежит пользователю
	ErrPostNotOwned = errors.New("post does not belong to user")

	// ErrPostsNotFound - посты пользователя не найдены
	ErrPostsNotFound = errors.New("user posts not found")

	// ErrEmptyTitle - пустой заголовок поста
	ErrEmptyTitle = errors.New("post title cannot be empty")

	// ErrEmptyContent - пустое содержимое поста
	ErrEmptyContent = errors.New("post content cannot be empty")

	// ErrInternalError - общая ошибка
	ErrInternalError = errors.New("internal error")
)
