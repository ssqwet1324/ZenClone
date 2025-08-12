package usecase

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
	GetLoginByUserID(ctx context.Context, id uuid.UUID) (string, error)
	UpdateRefreshToken(ctx context.Context, id uuid.UUID, refreshToken string) error
	GetRefreshTokenByUserID(ctx context.Context, id uuid.UUID) (*entity.RefreshTokenResponse, error)
	GetUserProfileByUsername(ctx context.Context, username string) (*entity.ProfileUserInfoResponse, error)
	UpdateUserProfile(ctx context.Context, id uuid.UUID, updateProfileInfo entity.UpdateUserProfileInfoRequest) error
	GetUserIdByUsername(ctx context.Context, username string) (*entity.UserResponse, error)
	SubscribeFromUser(ctx context.Context, followerID, followingID uuid.UUID) error
	GetSubsUser(ctx context.Context, userID uuid.UUID) (*entity.SubsList, error)
	UnsubscribeFromUser(ctx context.Context, followerID, followingID uuid.UUID) error
	UploadAvatar(ctx context.Context, userID uuid.UUID, bucketName string, avatarInfo entity.AvatarRequest) error
	GetAvatarURL(ctx context.Context, bucketName string, userID uuid.UUID) (string, error)
}
type UserService struct {
	repo RepositoryProvider
	cfg  *config.Config
	log  *zap.Logger
}

func New(repo RepositoryProvider, cfg *config.Config, log *zap.Logger) *UserService {
	return &UserService{
		repo: repo,
		cfg:  cfg,
		log:  log.Named("UserService"),
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

	if err != nil {
		return nil, fmt.Errorf("GetUserInfoByLogin error: %w", err)
	}

	compareData.ID = response.ID

	err = bcrypt.CompareHashAndPassword([]byte(response.Password), []byte(users.Password))
	if err != nil {
		return nil, fmt.Errorf("неверный пароль: %w", err)
	}

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

	return &tokenResponse, nil
}

// UpdateRefreshToken - обновляем refresh токен по ID
func (s *UserService) UpdateRefreshToken(ctx context.Context, req entity.UpdateRefreshTokenRequest) error {
	err := s.repo.UpdateRefreshToken(ctx, req.ID, req.RefreshToken)
	if err != nil {
		return fmt.Errorf("UpdateRefreshToken: error updating refresh token %w", err)
	}

	return nil
}

// GetUserProfileByUsername - получить информацию по профилю пользователя
func (s *UserService) GetUserProfileByUsername(ctx context.Context, username string) (*entity.ProfileUserInfoResponse, error) {
	userInfo, err := s.repo.GetUserProfileByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("GetUserProfileByUsername: error getting user info: %w", err)
	}

	avatarURL, err := s.GetAvatarURL(ctx, s.cfg.BucketName, username)
	if err != nil {
		return nil, fmt.Errorf("GetUserProfileByUsername: error getting avatar url: %w", err)
	}
	fmt.Println("URL:", avatarURL)

	return &entity.ProfileUserInfoResponse{
		FirstName:     userInfo.FirstName,
		LastName:      userInfo.LastName,
		Bio:           userInfo.Bio,
		UserAvatarUrl: avatarURL,
	}, nil
}

func (s *UserService) GetUserIDByUsername(ctx context.Context, username string) (*entity.UserResponse, error) {
	var user entity.UserResponse
	userID, err := s.repo.GetUserIdByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("GetUserIdByUsername error getting id by username: %w", err)
	}
	user.ID = userID.ID

	return &user, nil
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

// SubscribeToUser - подписаться на пользователя
func (s *UserService) SubscribeToUser(ctx context.Context, followerID uuid.UUID, username string) error {
	fmt.Println(followerID, username)
	followingID, err := s.repo.GetUserIdByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("GetUserIdByUsername error: %w", err)
	}
	fmt.Println(followerID, followingID)

	err = s.repo.SubscribeFromUser(ctx, followerID, followingID.ID)
	if err != nil {
		return fmt.Errorf("CreateSubToUser usecase  error: %w", err)
	}

	return nil
}

// GetSubsUser - получить подписки пользователя
func (s *UserService) GetSubsUser(ctx context.Context, username string) (*entity.SubsList, error) {
	targetUserID, err := s.repo.GetUserIdByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("GetUserIdByUsername error: %w", err)
	}

	data, err := s.repo.GetSubsUser(ctx, targetUserID.ID)
	if err != nil {
		return nil, fmt.Errorf("GetSubsUser error: %w", err)
	}

	if data == nil {
		return nil, fmt.Errorf("GetSubsUser data is nil")
	}

	return data, nil
}

// UnsubscribeFromUser - отписаться от пользователя
func (s *UserService) UnsubscribeFromUser(ctx context.Context, followerID uuid.UUID, targetUsername string) error {
	targetUserID, err := s.repo.GetUserIdByUsername(ctx, targetUsername)
	if err != nil {
		return fmt.Errorf("UnsubscribeFromUser error: %w", err)
	}

	err = s.repo.UnsubscribeFromUser(ctx, followerID, targetUserID.ID)
	if err != nil {
		return fmt.Errorf("UnsubscribeFromUser error: %w", err)
	}

	return nil
}

func (s *UserService) UploadAvatar(ctx context.Context, userID uuid.UUID, avatarInfo entity.AvatarRequest) error {
	err := s.repo.UploadAvatar(ctx, userID, s.cfg.BucketName, avatarInfo)
	if err != nil {
		return fmt.Errorf("UploadAvatar error: %w", err)
	}

	return nil
}

func (s *UserService) GetAvatarURL(ctx context.Context, bucketName string, username string) (string, error) {
	userID, err := s.repo.GetUserIdByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("GetAvatarURL error: %w", err)
	}

	url, err := s.repo.GetAvatarURL(ctx, bucketName, userID.ID)
	if err != nil {
		return "", fmt.Errorf("GetAvatarURL error: %w", err)
	}

	return url, nil
}
