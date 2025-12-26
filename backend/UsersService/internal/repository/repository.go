package repository

import (
	"UsersService/internal/config"
	"UsersService/internal/entity"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	defaultAvatar         = "defaultFoto/default.jpg"
	maxConnectionsFomPgx  = 20
	minConnectionsFromPgx = 5
)

// PostgresRepository - структура хранилищ
type PostgresRepository struct {
	db     *pgxpool.Pool
	client *minio.Client
	config *config.Config
}

// Init - инициализация repository
func Init(ctx context.Context, cfg *config.Config) (*PostgresRepository, error) {
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
	defer sqlDb.Close()

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

	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSl,
	})

	if err != nil {
		return nil, fmt.Errorf("PostgresRepository: error initializing MinIO client: %v", err)
	}

	return &PostgresRepository{
		db:     pool,
		client: minioClient,
		config: cfg,
	}, nil
}

// Close - закрытие бд
func (repo *PostgresRepository) Close() {
	repo.db.Close()
}

// AddUser - добавить пользователя в Бд
func (repo *PostgresRepository) AddUser(ctx context.Context, addUserInfo entity.AddUserRequest) error {
	_, err := repo.db.Exec(ctx, `INSERT INTO Users (id, login, password, username, first_name, last_name, bio) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		addUserInfo.ID,
		addUserInfo.Login,
		addUserInfo.Password,
		addUserInfo.Username,
		addUserInfo.FirstName,
		addUserInfo.LastName,
		addUserInfo.Bio,
	)
	if err != nil {
		return fmt.Errorf("AddUser: Error adding user in db: %v", err)
	}

	return nil
}

// CheckUser - проверяем существует ли пользователь
func (repo *PostgresRepository) CheckUser(ctx context.Context, username string) (bool, error) {
	var exists int

	err := repo.db.QueryRow(ctx, `SELECT 1 FROM users WHERE username=$1`, username).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetLoginByUserID - получаем логин по id пользователя
func (repo *PostgresRepository) GetLoginByUserID(ctx context.Context, id uuid.UUID) (string, error) {
	var login string
	err := repo.db.QueryRow(ctx, `SELECT login FROM Users WHERE id = $1`, id).Scan(&login)
	if err != nil {
		return "", fmt.Errorf("GetLoginByUserID: Error getting login by user: %v", err)
	}

	return login, nil
}

// GetUserInfoByLogin - получаем логин и пароль для ручки /compare-auth-data
func (repo *PostgresRepository) GetUserInfoByLogin(ctx context.Context, login string) (*entity.LoginResponse, error) {
	var userInfo entity.LoginResponse

	err := repo.db.QueryRow(ctx, `SELECT id, password FROM Users WHERE login = $1`, login).
		Scan(&userInfo.ID, &userInfo.Password)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("GetUserInfo: User not found")
	}

	if err != nil {
		return nil, fmt.Errorf("GetUserInfo: Error getting user information: %v", err)
	}

	return &userInfo, nil
}

// GetRefreshTokenByUserID - получаем refresh токен по id пользователя
func (repo *PostgresRepository) GetRefreshTokenByUserID(ctx context.Context, id uuid.UUID) (*entity.RefreshTokenResponse, error) {
	var refreshInfo entity.RefreshTokenResponse

	err := repo.db.QueryRow(ctx, `SELECT refresh_token FROM Users WHERE id = $1`, id).Scan(&refreshInfo.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("GetRefreshTokenByUserID: Error getting user information: %v", err)
	}

	return &refreshInfo, nil
}

// UpdateRefreshToken - обновляем refresh токен в БД
func (repo *PostgresRepository) UpdateRefreshToken(ctx context.Context, id uuid.UUID, refreshToken string) error {
	_, err := repo.db.Exec(ctx, `UPDATE Users SET refresh_token = $1 WHERE id = $2`, refreshToken, id)
	if err != nil {
		return fmt.Errorf("UpdateRefreshToken: error update refresh token %w", err)
	}

	return nil
}

// GetUserProfileByUsername - получаем данные пользователя для профиля
func (repo *PostgresRepository) GetUserProfileByUsername(ctx context.Context, username string) (*entity.ProfileUserInfoResponse, error) {
	var userInfoResponse entity.ProfileUserInfoResponse

	err := repo.db.QueryRow(ctx, `SELECT first_name, last_name, bio, avatar_url FROM Users WHERE username = $1`,
		username).Scan(&userInfoResponse.FirstName,
		&userInfoResponse.LastName,
		&userInfoResponse.Bio,
		&userInfoResponse.UserAvatarUrl)

	if err != nil {
		return nil, fmt.Errorf("GetUserProfileByUsername: Error getting user information: %v", err)
	}

	return &userInfoResponse, nil
}

// UpdateUserProfile обновить данные пользователя
func (repo *PostgresRepository) UpdateUserProfile(ctx context.Context, id uuid.UUID, updateProfileInfo entity.UpdateUserProfileInfoRequest) error {
	var setParts []string
	var args []interface{}
	argIdx := 1

	if updateProfileInfo.FirstName != nil {
		setParts = append(setParts, fmt.Sprintf("first_name = $%d", argIdx))
		args = append(args, *updateProfileInfo.FirstName)
		argIdx++
	}
	if updateProfileInfo.LastName != nil {
		setParts = append(setParts, fmt.Sprintf("last_name = $%d", argIdx))
		args = append(args, *updateProfileInfo.LastName)
		argIdx++
	}
	if updateProfileInfo.Bio != nil {
		setParts = append(setParts, fmt.Sprintf("bio = $%d", argIdx))
		args = append(args, *updateProfileInfo.Bio)
		argIdx++
	}
	if updateProfileInfo.PasswordNew != nil {
		setParts = append(setParts, fmt.Sprintf("password = $%d", argIdx))
		args = append(args, *updateProfileInfo.PasswordNew)
		argIdx++
	}

	if updateProfileInfo.Username != nil {
		setParts = append(setParts, fmt.Sprintf("username = $%d", argIdx))
		args = append(args, *updateProfileInfo.Username)
		argIdx++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("nothing to update")
	}

	args = append(args, id)

	query := fmt.Sprintf("UPDATE Users SET %s WHERE id = $%d", strings.Join(setParts, ", "), argIdx)

	_, err := repo.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("UpdateUserProfile: error updating data %w", err)
	}

	return nil
}

// GetUserIdByUsername - получить ID по username
func (repo *PostgresRepository) GetUserIdByUsername(ctx context.Context, username string) (*entity.UserResponse, error) {
	var userIDResponse entity.UserResponse
	err := repo.db.QueryRow(ctx, `SELECT id FROM users WHERE username = $1`, username).Scan(&userIDResponse.ID)
	if err != nil {
		return nil, fmt.Errorf("GetUserIdByUsername: Error getting user information: %v", err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("GetUserIdByUsername: User not found")
	}

	return &userIDResponse, nil
}

// SubscribeFromUser - подписаться на пользователя
func (repo *PostgresRepository) SubscribeFromUser(ctx context.Context, followerID, followingID uuid.UUID) error {
	rows, err := repo.db.Exec(ctx, `INSERT INTO subscriptions (follower_id, following_id) VALUES ($1, $2)`,
		followerID, followingID)
	if err != nil {
		return fmt.Errorf("CreateSubToUser: error inserting subscription: %w", err)
	}

	if rows.RowsAffected() != 1 {
		return fmt.Errorf("CreateSubToUser: expected to insert 1 row, but inserted %d", rows.RowsAffected())
	}

	return nil
}

// GetSubsUser - получить подписки пользователя
func (repo *PostgresRepository) GetSubsUser(ctx context.Context, userID uuid.UUID) (*entity.SubsList, error) {
	var subList entity.SubsList

	rows, err := repo.db.Query(ctx, `
		SELECT users.id, users.username, users.first_name, users.last_name
		FROM subscriptions
		JOIN users ON subscriptions.following_id = users.id
		WHERE subscriptions.follower_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("GetSubsUser: error executing query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sub entity.SubUserInfo
		if err := rows.Scan(&sub.ID, &sub.Username, &sub.FirstName, &sub.LastName); err != nil {
			return nil, fmt.Errorf("GetSubsUser: error scanning row: %w", err)
		}
		subList.Subs = append(subList.Subs, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetSubsUser: rows iteration error: %w", err)
	}

	return &subList, nil
}

// UnsubscribeFromUser - отписаться от пользователя
func (repo *PostgresRepository) UnsubscribeFromUser(ctx context.Context, followerID, followingID uuid.UUID) error {
	row, err := repo.db.Exec(ctx, `DELETE FROM subscriptions WHERE follower_id = $1 AND following_id = $2`, followerID, followingID)
	if err != nil {
		return fmt.Errorf("DeleteSub: error executing delete: %w", err)
	}

	if row.RowsAffected() == 0 {
		return fmt.Errorf("DeleteSub: no subscription found to delete")
	}

	return nil
}
