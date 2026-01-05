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
	GetUserProfileByID(ctx context.Context, userID uuid.UUID) (*entity.ProfileUserInfoResponse, error)
	UpdateUserProfile(ctx context.Context, id uuid.UUID, updateProfileInfo entity.UpdateUserProfileInfoRequest) error
	GetUserIdByUsername(ctx context.Context, username string) (string, error)
	CheckUser(ctx context.Context, username string) (bool, error)
	GetLoginByUserID(ctx context.Context, id uuid.UUID) (string, error)
	GlobalSearchPeople(ctx context.Context, firstName, lastName string, bucketName string) (*entity.PersonDateList, error)
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
	GetSubsUser(ctx context.Context, userID uuid.UUID, bucketName string) (*entity.SubsList, error)
	CheckSubscription(ctx context.Context, followerID string, followingID string) (bool, error)
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

// Usecase - структура бизнес логики
type Usecase struct {
	repo RepositoryProvider
	cfg  *config.Config
	log  *zap.Logger
}

// New - конструктор
func New(repo RepositoryProvider, cfg *config.Config, log *zap.Logger) *Usecase {
	return &Usecase{
		repo: repo,
		cfg:  cfg,
		log:  log.Named("Usecase"),
	}
}

// AddUser - создание/ добавление нового пользователя
func (s *Usecase) AddUser(ctx context.Context, addUserInfo entity.AddUserRequest) error {
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
func (s *Usecase) CompareAuthData(ctx context.Context, users entity.AuthRequest) (*entity.CompareDataResponse, error) {
	var compareData entity.CompareDataResponse

	response, err := s.repo.GetUserInfoByLogin(ctx, users.Login)
	if err != nil {
		s.log.Error("CompareAuthData: failed to get user info by login", zap.Error(err))
		return nil, entity.ErrUserNotFound
	}

	compareData.ID = response.ID
	compareData.Username = response.Username

	err = bcrypt.CompareHashAndPassword([]byte(response.Password), []byte(users.Password))
	if err != nil {
		s.log.Error("CompareAuthData: failed to compare password", zap.Error(err))
		return nil, entity.ErrIncorrectPassword
	}

	return &compareData, nil
}

// GetRefreshToken - получить refresh токен пользователя
func (s *Usecase) GetRefreshToken(ctx context.Context, id uuid.UUID) (*entity.TokenResponse, error) {
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
func (s *Usecase) UpdateRefreshToken(ctx context.Context, req entity.UpdateRefreshTokenRequest) error {
	err := s.repo.UpdateRefreshToken(ctx, req.ID, req.RefreshToken)
	if err != nil {
		s.log.Error("UpdateRefreshToken: failed to update refresh token", zap.Error(err))
		return entity.ErrFailedToUpdateRefreshToken
	}
	return nil
}

// GetUserProfileByID - получить профиль по ID пользователя
func (s *Usecase) GetUserProfileByID(ctx context.Context, yourUserID uuid.UUID, otherUserID uuid.UUID) (*entity.ProfileUserInfoResponse, error) {
	userInfo, err := s.repo.GetUserProfileByID(ctx, otherUserID)
	if err != nil {
		s.log.Error("GetUserProfileByID: failed to get user info by id", zap.String("userID", otherUserID.String()), zap.Error(err))
		return nil, entity.ErrFailedToGetUserInfo
	}

	avatarURL, err := s.repo.GetAvatarURL(ctx, s.cfg.BucketName, otherUserID)
	if err != nil {
		s.log.Error("GetUserProfileByID: failed to get avatar url", zap.Error(err))
		return nil, entity.ErrFailedToGetAvatarURL
	}

	isSubscribed, err := s.repo.CheckSubscription(ctx, yourUserID.String(), otherUserID.String())
	if err != nil {
		s.log.Error("GetUserProfileByID: failed to check subscription", zap.String("userID", otherUserID.String()), zap.Error(err))
		return nil, entity.ErrUserNotFound
	}

	return &entity.ProfileUserInfoResponse{
		FirstName:     userInfo.FirstName,
		LastName:      userInfo.LastName,
		Bio:           userInfo.Bio,
		UserAvatarUrl: avatarURL,
		IsSubscribed:  isSubscribed,
	}, nil
}

// UpdateUserProfile - обновить профиль
func (s *Usecase) UpdateUserProfile(ctx context.Context, id uuid.UUID, updateProfileInfo entity.UpdateUserProfileInfoRequest) (*entity.UpdateUserProfileInfoResponse, error) {
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
func (s *Usecase) GetUserIDByUsername(ctx context.Context, username string) (string, error) {
	userID, err := s.repo.GetUserIdByUsername(ctx, username)
	if err != nil {
		s.log.Error("GetUserIDByUsername: failed to get user info by username", zap.String("username", username), zap.Error(err))
		return "", entity.ErrInternalServer
	}

	if userID == "" {
		return "", entity.ErrUserNotFound
	}

	return userID, nil
}

// SubscribeToUser - подписаться на пользователя
func (s *Usecase) SubscribeToUser(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) error {
	err := s.repo.SubscribeFromUser(ctx, followerID, followingID)
	if err != nil {
		s.log.Error("SubscribeToUser: failed subscribe to user", zap.Error(err))
		return entity.ErrFailedToSubscribe
	}

	return nil
}

// GetSubsUser - получить подписки пользователя
func (s *Usecase) GetSubsUser(ctx context.Context, userID uuid.UUID) (*entity.SubsList, error) {
	data, err := s.repo.GetSubsUser(ctx, userID, s.cfg.BucketName)
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
func (s *Usecase) UnsubscribeFromUser(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) error {
	err := s.repo.UnsubscribeFromUser(ctx, followerID, followingID)
	if err != nil {
		s.log.Error("UnsubscribeFromUser: failed to unsubscribe from user", zap.Error(err))
		return entity.ErrFailedToUnsubscribe
	}

	return nil
}

// UploadAvatar - загрузить аватарку
func (s *Usecase) UploadAvatar(ctx context.Context, userID uuid.UUID, avatarInfo entity.AvatarRequest) error {
	err := s.repo.UploadAvatar(ctx, userID, s.cfg.BucketName, avatarInfo)
	if err != nil {
		s.log.Error("UploadAvatar: failed to upload avatar", zap.Error(err))
		return entity.ErrFailedToUploadAvatar
	}

	return nil
}

// GetAvatarURL - получить url аватара (метод оставлен для обратной совместимости, но теперь не используется в GetUserProfileByID)
func (s *Usecase) GetAvatarURL(ctx context.Context, bucketName string, userID uuid.UUID) (string, error) {
	url, err := s.repo.GetAvatarURL(ctx, bucketName, userID)
	if err != nil {
		s.log.Error("GetAvatarURL: failed to get avatar url", zap.Error(err))
		return "", entity.ErrFailedToGetAvatarURL
	}

	return url, nil
}

// GlobalSearchPeople - поиск профиля по имени фамилии
func (s *Usecase) GlobalSearchPeople(ctx context.Context, firstName, lastName string) (*entity.PersonDateList, error) {
	data, err := s.repo.GlobalSearchPeople(ctx, firstName, lastName, s.cfg.BucketName)
	if err != nil {
		s.log.Error("GlobalSearchPeople: failed to search people", zap.Error(err))
		return nil, err
	}

	if len(data.Persons) == 0 {
		return nil, entity.ErrUserNotFound
	}

	return data, nil
}
