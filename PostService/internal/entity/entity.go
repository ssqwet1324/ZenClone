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
	Posts []PostResponse `json:"posts"`
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
