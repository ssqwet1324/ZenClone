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

	return &PostgresRepository{dbPool}, nil
}

// AddUser - добавить пользователя в Бд(для ручки /add-user
func (repo *PostgresRepository) AddUser(ctx context.Context, addUserInfo entity.AddUserRequest) error {
	_, err := repo.DB.Exec(ctx, `INSERT INTO Users (id, login, password) VALUES ($1, $2, $3)`, addUserInfo.ID, addUserInfo.Login, addUserInfo.Password)
	if err != nil {
		return fmt.Errorf("AddUser: Error adding user in DB: %v", err)
	}

	fmt.Println("Repository AddUser:", addUserInfo.Login, addUserInfo.Password)

	//todo проверить на дубликат пользователя

	return nil
}

// GetUserInfoByLogin - получаем логин и пароль для ручки /compare-auth-data
func (repo *PostgresRepository) GetUserInfoByLogin(ctx context.Context, login string) (*entity.LoginResponse, error) {
	var userInfo entity.LoginResponse

	//!!!!!!!!!!!!!!!!!!!!!!!!!!
	err := repo.DB.QueryRow(ctx, `SELECT id, password FROM Users WHERE login= $1`, login).Scan(&userInfo.ID, &userInfo.Password)
	if err != nil {
		return nil, fmt.Errorf("GetUserInfo: Error getting user information: %v", err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("GetUserInfo: User not found")
	}

	return &userInfo, nil
}

// GetUserInfoByUserID - получаем refresh токен по id пользователя
func (repo *PostgresRepository) GetUserInfoByUserID(ctx context.Context, id uuid.UUID) (*entity.RefreshTokenResponse, error) {
	var refreshInfo entity.RefreshTokenResponse

	err := repo.DB.QueryRow(ctx, `SELECT refresh_token FROM Users WHERE id = $1`, id).Scan(&refreshInfo.RefreshToken)
	fmt.Println("Repository GetUserInfoByUserID:", id, refreshInfo.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("GetUserInfoByUserID: Error getting user information: %v", err)
	}

	return &refreshInfo, nil
}

// UpdateRefreshToken - обновляем refresh токен в БД
func (repo *PostgresRepository) UpdateRefreshToken(ctx context.Context, id uuid.UUID, refreshToken string) error {
	fmt.Println("REPO", "id:", id, "refreshToken:", refreshToken)

	_, err := repo.DB.Exec(ctx, `UPDATE Users SET refresh_token = $1 WHERE id = $2`, refreshToken, id)
	if err != nil {
		return fmt.Errorf("UpdateRefreshToken: error update refresh token %w", err)
	}

	return nil
}
