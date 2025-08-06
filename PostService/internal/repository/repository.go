package repository

import (
	"PostService/internal/config"
	"PostService/internal/entity"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

type PostgresRepository struct {
	DB  *pgxpool.Pool
	log *zap.Logger
	cfg *config.Config
}

func New(cfg *config.Config, log *zap.Logger) (*PostgresRepository, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DbUser,
		cfg.DbPassword,
		cfg.DbHost,
		strconv.Itoa(cfg.DbPort),
		cfg.DbName,
	)

	dbPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("PostgresRepository: Error connecrtion from pgxpool: %v", err)
	}

	return &PostgresRepository{
		DB:  dbPool,
		log: log.Named("PostRepository"),
		cfg: cfg,
	}, nil
}

// CreatePost - создать пост
func (repo *PostgresRepository) CreatePost(ctx context.Context, createPost entity.CreatePostRequest) (*entity.CreatePostResponse, error) {
	var postResponse entity.CreatePostResponse
	err := repo.DB.QueryRow(ctx,
		`INSERT INTO Posts (post_id, title, content, author_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING post_id, title, content, author_id, created_at`,
		createPost.ID,
		createPost.Title,
		createPost.Content,
		createPost.AuthorID,
	).Scan(&postResponse.ID, &postResponse.Title, &postResponse.Content, &postResponse.AuthorID, &postResponse.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("CreatePost: error inserting post: %v", err)
	}

	repo.log.Info("CreatePost successful", zap.String("post_id", postResponse.ID.String()))

	return &postResponse, nil
}

// UpdatePost - редактирование поста с помощью динамического запроса к бд
func (repo *PostgresRepository) UpdatePost(ctx context.Context, postID uuid.UUID, authorID uuid.UUID, updateReq entity.UpdateUserPostRequest) (*entity.UpdateUserPostResponse, error) {
	var postResponse entity.UpdateUserPostResponse
	var setParts []string
	var args []interface{}
	argIdx := 1

	if updateReq.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *updateReq.Title)
		argIdx++
	}

	if updateReq.Content != nil {
		setParts = append(setParts, fmt.Sprintf("content = $%d", argIdx))
		args = append(args, *updateReq.Content)
		argIdx++
	}

	if len(setParts) == 0 {
		return &postResponse, fmt.Errorf("nothing to update")
	}

	setParts = append(setParts, fmt.Sprintf(`updated_at = $%d`, argIdx))
	args = append(args, time.Now())
	argIdx++

	// postID и authorID для условия WHERE
	args = append(args, postID)
	postIDArgIdx := argIdx
	argIdx++

	args = append(args, authorID)
	authorIDArgIdx := argIdx

	query := fmt.Sprintf(
		`UPDATE Posts SET %s WHERE post_id = $%d AND author_id = $%d RETURNING title, content, updated_at`,
		strings.Join(setParts, ", "),
		postIDArgIdx,
		authorIDArgIdx,
	)

	err := repo.DB.QueryRow(ctx, query, args...).Scan(&postResponse.Title, &postResponse.Content, &postResponse.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &postResponse, fmt.Errorf("UpdatePost: post not found or not author")
		}

		return &postResponse, fmt.Errorf("UpdatePost: error updating post: %w", err)
	}

	repo.log.Info("UpdatePost successful", zap.String("post_id", postID.String()))

	return &postResponse, nil
}

// DeletePost - удалить пост
func (repo *PostgresRepository) DeletePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error {
	res, err := repo.DB.Exec(ctx, `DELETE FROM Posts WHERE post_id = $1 AND author_id = $2`, postID, userID)
	if err != nil {
		return fmt.Errorf("DeletePost: error deleting post: %v", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("DeletePost: no such post: %v", postID)
	}

	repo.log.Info("DeletePost: post deleted", zap.String("postID", postID.String()))

	return nil
}

func (repo *PostgresRepository) GetPostsUser(ctx context.Context, authorID uuid.UUID) (*entity.PostListResponse, error) {
	var postList entity.PostListResponse

	rows, err := repo.DB.Query(ctx, `SELECT post_id, title, content, created_at, updated_at FROM Posts WHERE author_id = $1`, authorID)
	if err != nil {
		return nil, fmt.Errorf("GetPostsUser: error getting posts: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var post entity.PostResponse
		err = rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("GetPostsUser: error scanning posts: %v", err)
		}

		postList.Posts = append(postList.Posts, post)
	}

	repo.log.Info("GetPostsUser: successful getting posts", zap.String("author_id", authorID.String()))

	return &postList, nil
}
