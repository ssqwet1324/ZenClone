package PostClient

import (
	"github.com/google/uuid"
	"time"
)

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
