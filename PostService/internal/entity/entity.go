package entity

import (
	"github.com/google/uuid"
	"time"
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
