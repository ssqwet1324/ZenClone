package usecase

import (
	"UsersService/internal/client/PostClient"
	"UsersService/internal/config"
	"UsersService/internal/entity"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type RepositoryProvider interface {
	AddUser(ctx context.Context, addUserInfo entity.AddUserRequest) error
	GetUserInfoByLogin(ctx context.Context, login string) (*entity.LoginResponse, error)
	GetLoginByUserID(ctx context.Context, id uuid.UUID) (string, error)
	UpdateRefreshToken(ctx context.Context, id uuid.UUID, refreshToken string) error
	GetRefreshTokenByUserID(ctx context.Context, id uuid.UUID) (*entity.RefreshTokenResponse, error)
	GetUserProfileByUsername(ctx context.Context, username string) (*entity.ProfileUserInfoResponse, error)
	UpdateUserProfile(ctx context.Context, id uuid.UUID, updateProfileInfo entity.UpdateUserProfileInfoRequest) error
}
type UserService struct {
	log    *zap.Logger
	repo   RepositoryProvider
	cfg    *config.Config
	client PostClient.ClientProvider
}

func New(log *zap.Logger, repo RepositoryProvider, cfg *config.Config, client PostClient.ClientProvider) *UserService {
	return &UserService{
		log:    log.Named("usecase"),
		repo:   repo,
		cfg:    cfg,
		client: client,
	}
}

func (s *UserService) AddUser(ctx context.Context, addUserInfo entity.AddUserRequest) error {
	err := s.repo.AddUser(ctx, addUserInfo)
	if err != nil {
		return fmt.Errorf("AddUser: error add login and hash %w", err)
	}

	s.log.Info("AddUser: Success add user")

	return nil
}

// CompareAuthData - тут надо сравнить hash пароля
func (s *UserService) CompareAuthData(ctx context.Context, users entity.AuthRequest) (*entity.CompareDataResponse, error) {
	var compareData entity.CompareDataResponse

	response, err := s.repo.GetUserInfoByLogin(ctx, users.Login)

	if err != nil {
		return nil, fmt.Errorf("GetUserInfoByLogin error: %w", err)
	}

	compareData.ID = response.ID

	err = bcrypt.CompareHashAndPassword([]byte(response.Password), []byte(users.Password))
	if err != nil {
		s.log.Info("CompareAuthData error", zap.Error(err))
		return nil, fmt.Errorf("неверный пароль: %w", err)
	}

	s.log.Info("CompareAuthData success compare auth data")

	return &compareData, nil
}

// GetRefreshToken - тут получаем refresh токен из БД
func (s *UserService) GetRefreshToken(ctx context.Context, id uuid.UUID) (*entity.TokenResponse, error) {
	var tokenResponse entity.TokenResponse

	response, err := s.repo.GetRefreshTokenByUserID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetRefreshTokenByUserID error: %w", err)
	}

	tokenResponse.RefreshToken = response.RefreshToken

	s.log.Info("GetRefreshToken: success get refresh token")

	return &tokenResponse, nil
}

// UpdateRefreshToken - обновляем refresh токен по ID
func (s *UserService) UpdateRefreshToken(ctx context.Context, req entity.UpdateRefreshTokenRequest) error {
	err := s.repo.UpdateRefreshToken(ctx, req.ID, req.RefreshToken)
	if err != nil {
		return fmt.Errorf("UpdateRefreshToken: error updating refresh token %w", err)
	}

	s.log.Info("UpdateRefreshToken: success update refresh token")

	return nil
}

// GetUserProfileByUsername - получить информацию по профилю пользователя
func (s *UserService) GetUserProfileByUsername(ctx context.Context, username string) (*entity.ProfileUserInfoResponse, error) {
	var profileInfo entity.ProfileUserInfoResponse

	info, err := s.repo.GetUserProfileByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("GetUserProfileByUsername error: %w", err)
	}

	profileInfo.FirstName = info.FirstName
	profileInfo.LastName = info.LastName
	profileInfo.Bio = info.Bio

	s.log.Info("GetUserProfileByUsername: success get user profile")

	return &profileInfo, nil
}

// GetPostsByUsername - получить посты пользователя в его профиле
func (s *UserService) GetPostsByUsername(ctx *gin.Context, usename string) (*entity.UserPosts, error) {
	var posts entity.UserPosts
	data, err := s.client.GetPostsUser(ctx, usename)
	if err != nil {
		s.log.Error("GetPostsByUsername: GetPostsByUsername error", zap.Error(err))
	}

	posts.Posts = data

	return &posts, nil
}

func (s *UserService) UpdateUserProfile(ctx context.Context, id uuid.UUID, updateProfileInfo entity.UpdateUserProfileInfoRequest) (*entity.UpdateUserProfileInfoResponse, error) {

	login, err := s.repo.GetLoginByUserID(ctx, id)

	// тут сверяем пароли
	if updateProfileInfo.PasswordNew != nil && updateProfileInfo.PasswordOld != nil {
		authRequest := entity.AuthRequest{
			Login:    login,
			Password: *updateProfileInfo.PasswordOld,
		}

		_, err := s.CompareAuthData(ctx, authRequest)
		if err != nil {
			return nil, fmt.Errorf("UpdateUserProfile CompareAuthData error: incorrect data: %w", err)
		}

		// Генерируем новый хеш
		newHashPassword, err := bcrypt.GenerateFromPassword([]byte(*updateProfileInfo.PasswordNew), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("UpdateUserProfile bcrypt error generating new hash: %w", err)
		}

		// меняем на хэш
		hashStr := string(newHashPassword)
		updateProfileInfo.PasswordNew = &hashStr
	}

	// Получаем ID пользователя
	userData, err := s.repo.GetUserInfoByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("UpdateUserProfile error getting user info: %w", err)
	}

	// обновляем данные
	err = s.repo.UpdateUserProfile(ctx, userData.ID, updateProfileInfo)
	if err != nil {
		return nil, fmt.Errorf("UpdateUserProfile error updating in repo: %w", err)
	}

	// ответ
	response := entity.UpdateUserProfileInfoResponse{
		Username:    updateProfileInfo.Username,
		FirstName:   updateProfileInfo.FirstName,
		LastName:    updateProfileInfo.LastName,
		Bio:         updateProfileInfo.Bio,
		PasswordNew: updateProfileInfo.PasswordNew,
	}

	return &response, nil
}
