package usecase

import (
	"UsersService/internal/config"
	"UsersService/internal/entity"
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UserRepo - описывает функции для пользователей
type UserRepo interface {
	AddUser(ctx context.Context, addUserInfo entity.AddUserRequest) error
	GetUserInfoByLogin(ctx context.Context, login string) (*entity.LoginResponse, error)
	GetUserProfileByUsername(ctx context.Context, username string) (*entity.ProfileUserInfoResponse, error)
	UpdateUserProfile(ctx context.Context, id uuid.UUID, updateProfileInfo entity.UpdateUserProfileInfoRequest) error
	GetUserIdByUsername(ctx context.Context, username string) (*entity.UserResponse, error)
	CheckUser(ctx context.Context, username string) (bool, error)
	GetLoginByUserID(ctx context.Context, id uuid.UUID) (string, error)
}

// TokenRepo - описывает функции взаимодействия с токенами
type TokenRepo interface {
	UpdateRefreshToken(ctx context.Context, id uuid.UUID, refreshToken string) error
	GetRefreshTokenByUserID(ctx context.Context, id uuid.UUID) (*entity.RefreshTokenResponse, error)
}

// SubscriptionRepo - описывает функции для взаимодейтсвия с подписками
type SubscriptionRepo interface {
	SubscribeFromUser(ctx context.Context, followerID, followingID uuid.UUID) error
	UnsubscribeFromUser(ctx context.Context, followerID, followingID uuid.UUID) error
	GetSubsUser(ctx context.Context, userID uuid.UUID) (*entity.SubsList, error)
}

// AvatarStorage - описывает функции хранения аватарок пользователекй
type AvatarStorage interface {
	GetAvatarURL(ctx context.Context, bucketName string, userID uuid.UUID) (string, error)
	UploadAvatar(ctx context.Context, userID uuid.UUID, bucketName string, avatarInfo entity.AvatarRequest) error
}

// RepositoryProvider - описывает все функции repository
type RepositoryProvider interface {
	UserRepo
	TokenRepo
	SubscriptionRepo
	AvatarStorage
}

// UserService - структура бизнес логики
type UserService struct {
	repo RepositoryProvider
	cfg  *config.Config
	log  *zap.Logger
}

// New - конструктор
func New(repo RepositoryProvider, cfg *config.Config, log *zap.Logger) *UserService {
	return &UserService{
		repo: repo,
		cfg:  cfg,
		log:  log.Named("UserService"),
	}
}

// AddUser - создание/ добавление нового пользователя
func (s *UserService) AddUser(ctx context.Context, addUserInfo entity.AddUserRequest) error {
	exist, err := s.repo.CheckUser(ctx, addUserInfo.Username)
	if err != nil {
		s.log.Error("AddUser: Error checking user", zap.String("username", addUserInfo.Username), zap.Error(err))
		return entity.ErrInternalServer
	}
	if exist {
		return entity.ErrUserAlreadyExists
	}

	err = s.repo.AddUser(ctx, addUserInfo)
	if err != nil {
		s.log.Error("AddUser: failed to add user", zap.Error(err))
		return entity.ErrFailedToAddUser
	}

	return nil
}

// CompareAuthData - сравнение данных
func (s *UserService) CompareAuthData(ctx context.Context, users entity.AuthRequest) (*entity.CompareDataResponse, error) {
	var compareData entity.CompareDataResponse

	response, err := s.repo.GetUserInfoByLogin(ctx, users.Login)
	if err != nil {
		s.log.Error("CompareAuthData: failed to get user info by login", zap.Error(err))
		return nil, entity.ErrUserNotFound
	}

	compareData.ID = response.ID

	err = bcrypt.CompareHashAndPassword([]byte(response.Password), []byte(users.Password))
	if err != nil {
		s.log.Error("CompareAuthData: failed to compare password", zap.Error(err))
		return nil, entity.ErrIncorrectPassword
	}

	return &compareData, nil
}

// GetRefreshToken - получить refresh токен пользователя
func (s *UserService) GetRefreshToken(ctx context.Context, id uuid.UUID) (*entity.TokenResponse, error) {
	var tokenResponse entity.TokenResponse

	response, err := s.repo.GetRefreshTokenByUserID(ctx, id)
	if err != nil {
		s.log.Error("GetRefreshToken: failed to get refresh token by userID", zap.Error(err))
		return nil, entity.ErrFailedToGetRefreshToken
	}

	tokenResponse.RefreshToken = response.RefreshToken

	return &tokenResponse, nil
}

// UpdateRefreshToken - обновить токен
func (s *UserService) UpdateRefreshToken(ctx context.Context, req entity.UpdateRefreshTokenRequest) error {
	err := s.repo.UpdateRefreshToken(ctx, req.ID, req.RefreshToken)
	if err != nil {
		s.log.Error("UpdateRefreshToken: failed to update refresh token", zap.Error(err))
		return entity.ErrFailedToUpdateRefreshToken
	}
	return nil
}

// GetUserProfileByUsername - получить профиль по username
func (s *UserService) GetUserProfileByUsername(ctx context.Context, username string) (*entity.ProfileUserInfoResponse, error) {
	userInfo, err := s.repo.GetUserProfileByUsername(ctx, username)
	if err != nil {
		s.log.Error("GetUserProfileByUsername: failed to get user info by username", zap.Error(err))
		return nil, entity.ErrFailedToGetUserInfo
	}

	avatarURL, err := s.GetAvatarURL(ctx, s.cfg.BucketName, username)
	if err != nil {
		s.log.Error("GetUserProfileByUsername: failed to get avatar url", zap.Error(err))
		return nil, entity.ErrFailedToGetAvatarURL
	}

	return &entity.ProfileUserInfoResponse{
		FirstName:     userInfo.FirstName,
		LastName:      userInfo.LastName,
		Bio:           userInfo.Bio,
		UserAvatarUrl: avatarURL,
	}, nil
}

// UpdateUserProfile - обновить профиль
func (s *UserService) UpdateUserProfile(ctx context.Context, id uuid.UUID, updateProfileInfo entity.UpdateUserProfileInfoRequest) (*entity.UpdateUserProfileInfoResponse, error) {
	login, err := s.repo.GetLoginByUserID(ctx, id)
	if err != nil {
		s.log.Error("UpdateUserProfile: failed to get user info by userID", zap.Error(err))
		return nil, entity.ErrFailedToGetUserInfo
	}

	if updateProfileInfo.PasswordNew != nil && updateProfileInfo.PasswordOld != nil {
		authRequest := entity.AuthRequest{
			Login:    login,
			Password: *updateProfileInfo.PasswordOld,
		}

		_, err := s.CompareAuthData(ctx, authRequest)
		if err != nil {
			s.log.Error("UpdateUserProfile: failed to compare auth data", zap.Error(err))
			return nil, entity.ErrIncorrectPassword
		}

		newHashPassword, err := bcrypt.GenerateFromPassword([]byte(*updateProfileInfo.PasswordNew), bcrypt.DefaultCost)
		if err != nil {
			s.log.Error("UpdateUserProfile: failed to hash password", zap.Error(err))
			return nil, entity.ErrInternalServer
		}

		hashStr := string(newHashPassword)
		updateProfileInfo.PasswordNew = &hashStr
	}

	userData, err := s.repo.GetUserInfoByLogin(ctx, login)
	if err != nil {
		s.log.Error("UpdateUserProfile: failed to get user info by login", zap.Error(err))
		return nil, entity.ErrFailedToGetUserInfo
	}

	err = s.repo.UpdateUserProfile(ctx, userData.ID, updateProfileInfo)
	if err != nil {
		s.log.Error("UpdateUserProfile: failed to update user profile", zap.Error(err))
		return nil, entity.ErrFailedToUpdateProfile
	}

	response := entity.UpdateUserProfileInfoResponse{
		Username:    updateProfileInfo.Username,
		FirstName:   updateProfileInfo.FirstName,
		LastName:    updateProfileInfo.LastName,
		Bio:         updateProfileInfo.Bio,
		PasswordNew: updateProfileInfo.PasswordNew,
	}

	return &response, nil
}

// GetUserIDByUsername - получить id пользователя по username
func (s *UserService) GetUserIDByUsername(ctx context.Context, username string) (*entity.UserResponse, error) {
	var user entity.UserResponse

	userID, err := s.repo.GetUserIdByUsername(ctx, username)
	if err != nil {
		s.log.Error("GetUserIDByUsername: failed to get user info by username", zap.String("username", username), zap.Error(err))
		return nil, entity.ErrInternalServer
	}

	if userID == nil {
		return nil, entity.ErrUserNotFound
	}

	user.ID = userID.ID

	return &user, nil
}

// SubscribeToUser - подписаться на пользователя
func (s *UserService) SubscribeToUser(ctx context.Context, followerID uuid.UUID, username string) error {
	followingID, err := s.repo.GetUserIdByUsername(ctx, username)
	if err != nil {
		s.log.Error("SubscribeToUser: failed get userID by username", zap.Error(err))
		return entity.ErrUserNotFound
	}

	err = s.repo.SubscribeFromUser(ctx, followerID, followingID.ID)
	if err != nil {
		s.log.Error("SubscribeToUser: failed subscribe to user", zap.Error(err))
		return entity.ErrFailedToSubscribe
	}

	return nil
}

// GetSubsUser - получить подписки пользователя
func (s *UserService) GetSubsUser(ctx context.Context, username string) (*entity.SubsList, error) {
	targetUserID, err := s.repo.GetUserIdByUsername(ctx, username)
	if err != nil {
		s.log.Error("GetSubsUser: failed to get user info by username", zap.Error(err))
		return nil, entity.ErrUserNotFound
	}

	data, err := s.repo.GetSubsUser(ctx, targetUserID.ID)
	if err != nil {
		s.log.Error("GetSubsUser: failed to get subs user", zap.Error(err))
		return nil, entity.ErrFailedToGetUserInfo
	}

	if data == nil {
		return nil, entity.ErrNoSubscriptions
	}

	return data, nil
}

// UnsubscribeFromUser - отписаться от пользователя
func (s *UserService) UnsubscribeFromUser(ctx context.Context, followerID uuid.UUID, targetUsername string) error {
	targetUserID, err := s.repo.GetUserIdByUsername(ctx, targetUsername)
	if err != nil {
		s.log.Error("UnsubscribeFromUser: failed to get user info by username", zap.Error(err))
		return entity.ErrUserNotFound
	}

	err = s.repo.UnsubscribeFromUser(ctx, followerID, targetUserID.ID)
	if err != nil {
		s.log.Error("UnsubscribeFromUser: failed to unsubscribe from user", zap.Error(err))
		return entity.ErrFailedToUnsubscribe
	}

	return nil
}

// UploadAvatar - загрузить аватарку
func (s *UserService) UploadAvatar(ctx context.Context, userID uuid.UUID, avatarInfo entity.AvatarRequest) error {
	err := s.repo.UploadAvatar(ctx, userID, s.cfg.BucketName, avatarInfo)
	if err != nil {
		s.log.Error("UploadAvatar: failed to upload avatar", zap.Error(err))
		return entity.ErrFailedToUploadAvatar
	}

	return nil
}

// GetAvatarURL - получить url аватара
func (s *UserService) GetAvatarURL(ctx context.Context, bucketName string, username string) (string, error) {
	userID, err := s.repo.GetUserIdByUsername(ctx, username)
	if err != nil {
		s.log.Error("GetAvatarURL: failed to get user info by username", zap.Error(err))
		return "", entity.ErrUserNotFound
	}

	url, err := s.repo.GetAvatarURL(ctx, bucketName, userID.ID)
	if err != nil {
		s.log.Error("GetAvatarURL: failed to get avatar url", zap.Error(err))
		return "", entity.ErrFailedToGetAvatarURL
	}

	return url, nil
}
