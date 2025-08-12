package repository

import (
	"UsersService/internal/config"
	"UsersService/internal/entity"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"strconv"
	"strings"
	"time"
)

const (
	minioEndpoint = "http://localhost:9000"
	defaultAvatar = "defaultFoto/default.jpg"
)

type PostgresRepository struct {
	DB     *pgxpool.Pool
	Client *minio.Client
}

func New(cfg *config.Config) (*PostgresRepository, error) {
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

	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSl,
	})
	if err != nil {
		return nil, fmt.Errorf("PostgresRepository: error initializing MinIO client: %v", err)
	}

	return &PostgresRepository{
		DB:     dbPool,
		Client: minioClient,
	}, nil
}

// AddUser - добавить пользователя в Бд(для ручки /add-user
func (repo *PostgresRepository) AddUser(ctx context.Context, addUserInfo entity.AddUserRequest) error {
	_, err := repo.DB.Exec(ctx, `INSERT INTO Users (id, login, password, username, first_name, last_name, bio) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		addUserInfo.ID,
		addUserInfo.Login,
		addUserInfo.Password,
		addUserInfo.Username,
		addUserInfo.FirstName,
		addUserInfo.LastName,
		addUserInfo.Bio,
	)
	if err != nil {
		return fmt.Errorf("AddUser: Error adding user in DB: %v", err)
	}

	return nil
}

// GetLoginByUserID - получаем логин по id пользователя
func (repo *PostgresRepository) GetLoginByUserID(ctx context.Context, id uuid.UUID) (string, error) {
	var login string
	err := repo.DB.QueryRow(ctx, `SELECT login FROM Users WHERE id = $1`, id).Scan(&login)
	if err != nil {
		return "", fmt.Errorf("GetLoginByUserID: Error getting login by user: %v", err)
	}

	return login, nil
}

// GetUserInfoByLogin - получаем логин и пароль для ручки /compare-auth-data
func (repo *PostgresRepository) GetUserInfoByLogin(ctx context.Context, login string) (*entity.LoginResponse, error) {
	var userInfo entity.LoginResponse

	err := repo.DB.QueryRow(ctx, `SELECT id, password FROM Users WHERE login = $1`, login).Scan(&userInfo.ID, &userInfo.Password)
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

	err := repo.DB.QueryRow(ctx, `SELECT refresh_token FROM Users WHERE id = $1`, id).Scan(&refreshInfo.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("GetRefreshTokenByUserID: Error getting user information: %v", err)
	}

	return &refreshInfo, nil
}

// UpdateRefreshToken - обновляем refresh токен в БД
func (repo *PostgresRepository) UpdateRefreshToken(ctx context.Context, id uuid.UUID, refreshToken string) error {
	_, err := repo.DB.Exec(ctx, `UPDATE Users SET refresh_token = $1 WHERE id = $2`, refreshToken, id)
	if err != nil {
		return fmt.Errorf("UpdateRefreshToken: error update refresh token %w", err)
	}

	return nil
}

// GetUserProfileByUsername - получаем данные пользователя для профиля
func (repo *PostgresRepository) GetUserProfileByUsername(ctx context.Context, username string) (*entity.ProfileUserInfoResponse, error) {
	var userInfoResponse entity.ProfileUserInfoResponse

	err := repo.DB.QueryRow(ctx, `SELECT first_name, last_name, bio, avatar_url FROM Users WHERE username = $1`, username).
		Scan(&userInfoResponse.FirstName,
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

	_, err := repo.DB.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("UpdateUserProfile: error updating data %w", err)
	}

	return nil
}

// GetUserIdByUsername - получить ID по username
func (repo *PostgresRepository) GetUserIdByUsername(ctx context.Context, username string) (*entity.UserResponse, error) {
	var userIDResponse entity.UserResponse
	err := repo.DB.QueryRow(ctx, `SELECT id FROM users WHERE username = $1`, username).Scan(&userIDResponse.ID)
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
	rows, err := repo.DB.Exec(ctx, `INSERT INTO subscriptions (follower_id, following_id) VALUES ($1, $2)`, followerID, followingID)
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

	rows, err := repo.DB.Query(ctx, `
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
	row, err := repo.DB.Exec(ctx, `DELETE FROM subscriptions WHERE follower_id = $1 AND following_id = $2`, followerID, followingID)
	if err != nil {
		return fmt.Errorf("DeleteSub: error executing delete: %w", err)
	}

	if row.RowsAffected() == 0 {
		return fmt.Errorf("DeleteSub: no subscription found to delete")
	}

	return nil
}

// UploadAvatar - загружаем фото и сохраняем его имя в бд
func (repo *PostgresRepository) UploadAvatar(ctx context.Context, userID uuid.UUID, bucketName string, avatarInfo entity.AvatarRequest) error {
	// timestamp, чтобы ссылка менялась при каждой загрузке
	objectName := fmt.Sprintf("%s/avatar_%d.jpg", userID.String(), time.Now().Unix())

	_, err := repo.Client.PutObject(ctx, bucketName, objectName, avatarInfo.Reader, avatarInfo.Size, minio.PutObjectOptions{
		ContentType: "image/jpg",
	})
	if err != nil {
		return fmt.Errorf("UploadAvatar: error uploading avatar: %w", err)
	}

	// Сохраняем путь к аватару в БД
	_, err = repo.DB.Exec(ctx, `UPDATE users SET avatar_url = $1 WHERE id = $2`, objectName, userID)
	if err != nil {
		return fmt.Errorf("UploadAvatar: error saving avatar_url in DB: %w", err)
	}

	return nil
}

// GetAvatarURL - получаем url аватарки по его имени
func (repo *PostgresRepository) GetAvatarURL(ctx context.Context, bucketName string, userID uuid.UUID) (string, error) {
	var objectName string
	err := repo.DB.QueryRow(ctx, `SELECT avatar_url FROM users WHERE id = $1`, userID).Scan(&objectName)
	if err != nil {
		return "", fmt.Errorf("GetAvatarURL: error fetching avatar_url from DB: %w", err)
	}

	if objectName == "default" {
		// Возвращаем дефолтную аватарку
		avatarURL := fmt.Sprintf(minioEndpoint+"/%s/%s", bucketName, defaultAvatar)

		return avatarURL, nil
	}

	avatarURL := fmt.Sprintf("%s/%s/%s", minioEndpoint, bucketName, objectName)

	return avatarURL, nil
}
