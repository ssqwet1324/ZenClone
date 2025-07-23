package service

import (
	"UsersService/internal/config"
	"UsersService/internal/entity"
	"context"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type RepositoryProvider interface {
	AddUser(ctx context.Context, addUserInfo entity.AddUserRequest) error
	GetUserInfoByLogin(ctx context.Context, login string) (*entity.LoginResponse, error)
	GetUserInfoByUserID(ctx context.Context, id uuid.UUID) (*entity.RefreshTokenResponse, error)
	UpdateRefreshToken(ctx context.Context, id uuid.UUID, refreshToken string) error
}
type UserService struct {
	log  *zap.Logger
	repo RepositoryProvider
	cfg  *config.Config
}

func New(log *zap.Logger, repo RepositoryProvider, cfg *config.Config) *UserService {
	return &UserService{
		log:  log,
		repo: repo,
		cfg:  cfg,
	}
}

func (s *UserService) AddUser(ctx context.Context, addUserInfo entity.AddUserRequest) error {
	err := s.repo.AddUser(ctx, addUserInfo)
	if err != nil {
		return fmt.Errorf("AddUser: error add login and hash %w", err)
	}

	return nil
}

// CompareAuthData - тут надо сравнить hash пароля
func (s *UserService) CompareAuthData(ctx context.Context, users entity.AuthRequest) (*entity.CompareDataResponse, error) {
	var compareData entity.CompareDataResponse

	response, err := s.repo.GetUserInfoByLogin(ctx, users.Login)

	//s.log.Info("CompareAuthData input passwords",
	//	zap.String("USERS: passwordFromUser", users.Password),
	//	zap.String("RESPONSE: passwordHashFromDB", response.Password),
	//)

	if err != nil {
		return nil, fmt.Errorf("GetUserInfoByLogin error: %w", err)
	}

	compareData.ID = response.ID

	err = bcrypt.CompareHashAndPassword([]byte(response.Password), []byte(users.Password))
	if err != nil {
		s.log.Info("CompareAuthData error", zap.Error(err))
		return nil, fmt.Errorf("неверный пароль: %w", err)
	}

	//fmt.Println("CompareAuthData service", "Byteresponse:", []byte(response.Password), "UsersPassword:", []byte(users.Password))

	////!!!!!!!!!!!!!!!!!!
	//fmt.Println("Service UsersService CompareAuthData", &compareData)

	return &compareData, nil
}

// GetRefreshToken - тут получаем refresh токен из БД
func (s *UserService) GetRefreshToken(ctx context.Context, id uuid.UUID) (*entity.TokenResponse, error) {
	var tokenResponse entity.TokenResponse

	response, err := s.repo.GetUserInfoByUserID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetUserInfoByUserID error: %w", err)
	}

	tokenResponse.RefreshToken = response.RefreshToken

	return &tokenResponse, nil
}

// UpdateRefreshToken - обновляем refresh токен по ID
func (s *UserService) UpdateRefreshToken(ctx context.Context, req entity.UpdateRefreshTokenRequest) error {
	fmt.Println("id:", req.ID, "token: ", req.RefreshToken)

	err := s.repo.UpdateRefreshToken(ctx, req.ID, req.RefreshToken)

	if err != nil {
		return fmt.Errorf("UpdateRefreshToken: error updating refresh token %w", err)
	}

	return nil
}
