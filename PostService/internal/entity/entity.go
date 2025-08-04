package entity

import (
	"github.com/google/uuid"
	"time"
)

type CreatePostRequest struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorID  string    `json:"author_id"` // Если что махнуть на username
	CreatedAt time.Time `json:"created_at"`
}

type CreatePostResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorID  string    `json:"author_id"`
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
