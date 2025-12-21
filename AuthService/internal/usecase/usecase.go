package usecase

import (
	"AuthService/internal/client/UsersClient"
	"AuthService/internal/config"
	"AuthService/internal/entity"
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenTTL = time.Hour
)

type RepositoryProvider interface {
	SaveRefreshToken(ctx context.Context, userID, refreshToken string) error
	GetRefreshToken(ctx context.Context, userID string) (string, error)
}
type UseCase struct {
	repo   RepositoryProvider
	log    *zap.Logger
	client UsersClient.ClientProvider
	cfg    *config.Config
}

func New(repo RepositoryProvider, log *zap.Logger, client UsersClient.ClientProvider, cfg *config.Config) *UseCase {
	return &UseCase{
		repo:   repo,
		log:    log,
		client: client,
		cfg:    cfg,
	}
}

// generateUserID - генерируем уникальный id пользователя
func generateUserID() string {
	return uuid.New().String()
}

// SaveRefreshToken имплементирующем интерфейс
func (s *UseCase) SaveRefreshToken(ctx context.Context, userID, refreshToken string) error {
	err := s.repo.SaveRefreshToken(ctx, userID, refreshToken)
	if err != nil {
		s.log.Error("Failed to save refresh token",
			zap.String("userID", userID),
			zap.Error(err),
		)
		return entity.ErrSaveRefreshToken
	}

	return nil
}

// GetRefreshToken - имплементирующем интерфейс
func (s *UseCase) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	// Пробуем получить из Redis
	refreshToken, err := s.repo.GetRefreshToken(ctx, userID)
	if err == nil && refreshToken != "" {
		return refreshToken, nil
	}

	// Если нет в Redis пробуем получить из UsersService
	token, err := s.client.GetRefreshToken(ctx, UsersClient.TokenRequest{ID: userID})
	if err != nil || token == "" {
		s.log.Error("Failed to get refresh token from Redis and UsersService",
			zap.String("userID", userID),
			zap.Error(err),
		)
		return "", entity.ErrGetRefreshToken
	}

	// Кэшируем обратно в Redis
	_ = s.repo.SaveRefreshToken(ctx, userID, token)

	return token, nil
}

// generateNewRefreshToken - создаем новый refresh токен
func generateNewRefreshToken() string {
	return uuid.NewString()
}

// GenerateAccessToken - создаем новый access(jwt) токен
func (s *UseCase) GenerateAccessToken(userID, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(accessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		s.log.Error("GenerateAccessToken: Failed to sign token",
			zap.String("userID", userID),
			zap.Error(err),
		)
		return "", entity.ErrSignToken
	}

	return tokenString, nil
}

// UpdateRefreshToken - обновляем устаревший refresh токен пользователя
func (s *UseCase) UpdateRefreshToken(ctx context.Context, userID string) (string, error) {
	newRefreshToken := generateNewRefreshToken()
	err := s.repo.SaveRefreshToken(ctx, userID, newRefreshToken)
	if err != nil {
		s.log.Error("UpdateRefreshToken: Failed to save refresh token in UpdateRefreshToken",
			zap.String("userID", userID),
			zap.Error(err),
		)
		return "", entity.ErrSaveRefreshToken
	}

	return newRefreshToken, nil
}

// ExtractUserIDFromToken взять userID из заголовка токена
func (s *UseCase) ExtractUserIDFromToken(tokenStr, secret string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		//проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			s.log.Error("ExtractUserIDFromToken: Unexpected signing method")
			return nil, entity.ErrUnexpectedSigningMethod
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		s.log.Error("ExtractUserIDFromToken: Invalid token",
			zap.Error(err),
		)
		return "", entity.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		s.log.Error("ExtractUserIDFromToken: Cannot parse claims")
		return "", entity.ErrCannotParseClaims
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		s.log.Error("ExtractUserIDFromToken: UserID not found in token claims")
		return "", entity.ErrUserIDNotFound
	}

	return userID, nil
}

// RegisterUser - регистрируем нового пользователя и посылаем данные в Users UseCase
func (s *UseCase) RegisterUser(ctx context.Context, reg entity.RegisterRequest) (string, string, string, error) {
	userID := generateUserID()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("RegisterUser: Failed to hash password",
			zap.String("login", reg.Login),
			zap.Error(err),
		)
		return "", "", "", entity.ErrHashPassword
	}

	err = s.client.AddUser(ctx, UsersClient.RegisterRequest{
		ID:        userID,
		Login:     reg.Login,
		Password:  string(hashedPassword),
		Username:  reg.Username,
		FirstName: reg.FirstName,
		LastName:  reg.LastName,
		Bio:       reg.Bio,
	})
	if err != nil {
		s.log.Error("RegisterUser: Failed to register user",
			zap.String("userID", userID),
			zap.String("login", reg.Login),
			zap.Error(err),
		)
		return "", "", "", entity.ErrRegisterUser
	}

	refreshToken := generateNewRefreshToken()
	err = s.repo.SaveRefreshToken(ctx, userID, refreshToken)
	if err != nil {
		s.log.Error("RegisterUser: Failed to save refresh token in RegisterUser",
			zap.String("userID", userID),
			zap.Error(err),
		)
		return "", "", "", entity.ErrSaveRefreshToken
	}

	accessToken, err := s.GenerateAccessToken(userID, s.cfg.JWTSecret)
	if err != nil {
		s.log.Error("RegisterUser: Failed to generate access token in RegisterUser",
			zap.String("userID", userID),
			zap.Error(err),
		)
		return "", "", "", entity.ErrGenerateAccessToken
	}

	return userID, accessToken, refreshToken, nil
}

// LoginAccount - входим в аккаунт и отдаем новые токены
func (s *UseCase) LoginAccount(ctx context.Context, login, password string) (string, string, string, error) {
	userID, err := s.client.CompareAuthData(ctx, UsersClient.AuthRequest{
		Login:    login,
		Password: password,
	})

	if err != nil {
		s.log.Error("LoginAccount: Failed to compare auth data",
			zap.String("login", login),
			zap.Error(err),
		)
		return "", "", "", entity.ErrCompareAuthData
	}

	//обновляем токен в redis
	newRefreshToken, err := s.UpdateRefreshToken(ctx, userID)
	if err != nil {
		s.log.Error("LoginAccount: Failed to update refresh token in LoginAccount",
			zap.String("userID", userID),
			zap.Error(err),
		)
		return "", "", "", entity.ErrUpdateRefreshToken
	}

	//обновляем токен в redis
	newAccessToken, err := s.GenerateAccessToken(userID, s.cfg.JWTSecret)
	if err != nil {
		s.log.Error("LoginAccount: Failed to generate access token in LoginAccount",
			zap.String("userID", userID),
			zap.Error(err),
		)
		return "", "", "", entity.ErrGenerateAccessToken
	}

	// обновляем токен через клиента UpdateRefreshToken
	var token UsersClient.UpdateRefreshTokenRequest

	token.ID = userID
	token.RefreshToken = newRefreshToken

	err = s.client.UpdateRefreshToken(ctx, token)

	return userID, newAccessToken, newRefreshToken, nil
}

// RefreshTokens - Отдаем новые токены
func (s *UseCase) RefreshTokens(ctx context.Context, refreshToken, authHeader string) (string, string, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		s.log.Error("RefreshTokens: Invalid authorization header")
		return "", "", entity.ErrInvalidAuthHeader
	}

	tokenString := parts[1]

	// Get userID from access token
	userID, err := s.ExtractUserIDFromToken(tokenString, s.cfg.JWTSecret)
	if err != nil {
		s.log.Error("RefreshTokens: Invalid access token in RefreshTokens",
			zap.Error(err),
		)
		return "", "", entity.ErrInvalidToken
	}

	// Получаем сохраненный refresh token
	storedRefreshToken, err := s.GetRefreshToken(ctx, userID)
	if err != nil || storedRefreshToken == "" {
		// Если не найден — генерируем новый, сохраняем в Redis и обновляем в UsersService
		newRefreshToken := generateNewRefreshToken()
		_ = s.SaveRefreshToken(ctx, userID, newRefreshToken)
		_ = s.client.UpdateRefreshToken(ctx, UsersClient.UpdateRefreshTokenRequest{
			ID:           userID,
			RefreshToken: newRefreshToken,
		})
		storedRefreshToken = newRefreshToken
	}

	// Сравниваем токены
	if storedRefreshToken != refreshToken {
		s.log.Error("RefreshTokens: Refresh token mismatch",
			zap.String("userID", userID),
		)
		return "", "", entity.ErrRefreshTokenMismatch
	}

	// Generate a new refresh token
	newRefreshToken := generateNewRefreshToken()
	if err := s.SaveRefreshToken(ctx, userID, newRefreshToken); err != nil {
		s.log.Error("RefreshTokens: Failed to save refresh token in RefreshTokens",
			zap.String("userID", userID),
			zap.Error(err),
		)
		return "", "", entity.ErrSaveRefreshToken
	}
	_ = s.client.UpdateRefreshToken(ctx, UsersClient.UpdateRefreshTokenRequest{
		ID:           userID,
		RefreshToken: newRefreshToken,
	})

	// Generate a new access token
	newAccessToken, err := s.GenerateAccessToken(userID, s.cfg.JWTSecret)
	if err != nil {
		s.log.Error("RefreshTokens: Failed to generate access token in RefreshTokens",
			zap.String("userID", userID),
			zap.Error(err),
		)
		return "", "", entity.ErrGenerateAccessToken
	}

	return newRefreshToken, newAccessToken, nil
}
