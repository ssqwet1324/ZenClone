package repository

import (
	"PostService/internal/config"
	"PostService/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

const (
	maxConnectionsFomPgx  = 20
	minConnectionsFromPgx = 5
)

type PostgresRepository struct {
	db  *pgxpool.Pool
	log *zap.Logger
	cfg *config.Config
}

// Init - инициализация repository
func Init(ctx context.Context, cfg *config.Config, log *zap.Logger) (*PostgresRepository, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.CreateDsn())
	if err != nil {
		return nil, fmt.Errorf("parse db config error: %w", err)
	}

	poolCfg.MaxConns = maxConnectionsFomPgx
	poolCfg.MinConns = minConnectionsFromPgx

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("db connection error: %w", err)
	}

	// проверка соединения
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db ping error: %w", err)
	}
	sqlDb := stdlib.OpenDB(*pool.Config().ConnConfig)
	defer func(sqlDb *sql.DB) {
		err := sqlDb.Close()
		if err != nil {
			log.Warn("PostgresRepository: Error closing PostService", zap.Error(err))
		}
	}(sqlDb)

	driver, err := postgres.WithInstance(sqlDb, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"pgx", driver,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create migrate: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("Migration initialized successfully")

	return &PostgresRepository{
		db:  pool,
		log: log.Named("Repository"),
		cfg: cfg,
	}, nil
}

// Close - закрытие бд
func (repo *PostgresRepository) Close() {
	repo.db.Close()
}

// CreatePost - создать пост
func (repo *PostgresRepository) CreatePost(ctx context.Context, createPost entity.CreatePostResponse) (*entity.CreatePostResponse, error) {
	var postResponse entity.CreatePostResponse

	err := repo.db.QueryRow(ctx,
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

	// postID и authorID
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

	err := repo.db.QueryRow(ctx, query, args...).Scan(&postResponse.Title, &postResponse.Content, &postResponse.UpdatedAt)
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
	res, err := repo.db.Exec(ctx, `DELETE FROM Posts WHERE post_id = $1 AND author_id = $2`, postID, userID)
	if err != nil {
		return fmt.Errorf("DeletePost: error deleting post: %v", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("DeletePost: no such post: %v", postID)
	}

	repo.log.Info("DeletePost: post deleted", zap.String("postID", postID.String()))

	return nil
}

// GetPostsUser - получить посты пользователя
func (repo *PostgresRepository) GetPostsUser(ctx context.Context, authorID uuid.UUID, limit int, cursor *entity.PostCursor) (*entity.PostListResponse, error) {
	var postList entity.PostListResponse
	var rows pgx.Rows
	var err error

	// делаем сначала обынчный запрос на limit, далее уже с cursor
	if cursor == nil {
		rows, err = repo.db.Query(ctx, `SELECT post_id, title, content, created_at, updated_at FROM posts
            WHERE author_id = $1 ORDER BY created_at DESC, post_id DESC LIMIT $2`, authorID, limit)
	} else {
		rows, err = repo.db.Query(ctx, `SELECT post_id, title, content, created_at, updated_at FROM posts
            WHERE author_id = $1 AND (created_at < $2 OR (created_at = $2 AND post_id < $3))
            ORDER BY created_at DESC, post_id DESC LIMIT $4`, authorID, cursor.CreatedAt, cursor.ID, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("GetPostsUser: error getting posts: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var post entity.PostResponse
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetPostsUser: scan error: %w", err)
		}
		postList.Posts = append(postList.Posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetPostsUser: error iterating rows: %v", err)
	}

	if len(postList.Posts) > 0 {
		last := postList.Posts[len(postList.Posts)-1]
		postList.NextCursor = &entity.PostCursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		}
	}

	return &postList, nil
}
