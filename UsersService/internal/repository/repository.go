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
	"strconv"
	"strings"
)

type PostgresRepository struct {
	DB *pgxpool.Pool
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

	return &PostgresRepository{
		DB: dbPool,
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
		addUserInfo.Bio, //убрать надо будет т.к чтобы о себе сразу не писать нечего
	)
	if err != nil {
		return fmt.Errorf("AddUser: Error adding user in DB: %v", err)
	}

	return nil
}

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

	err := repo.DB.QueryRow(ctx, `SELECT id, password FROM Users WHERE login= $1`, login).Scan(&userInfo.ID, &userInfo.Password)
	if err != nil {
		return nil, fmt.Errorf("GetUserInfo: Error getting user information: %v", err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("GetUserInfo: User not found")
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

	err := repo.DB.QueryRow(ctx, `SELECT first_name, last_name, bio FROM Users WHERE username = $1`, username).
		Scan(&userInfoResponse.FirstName,
			&userInfoResponse.LastName,
			&userInfoResponse.Bio)

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
