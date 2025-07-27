package service

import (
	"AuthService/internal/client/UsersClient"
	"AuthService/internal/config"
	"AuthService/internal/entity"
	"context"
	"fmt"
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
type Service struct {
	repo   RepositoryProvider
	log    *zap.Logger
	client UsersClient.ClientProvider
	cfg    *config.Config
}

func New(repo RepositoryProvider, log *zap.Logger, client UsersClient.ClientProvider, cfg *config.Config) *Service {
	return &Service{
		repo:   repo,
		log:    log,
		client: client,
		cfg:    cfg,
	}
}

// GenerateUserID - генерируем уникальный id пользователя
func (s *Service) GenerateUserID() string {
	return uuid.New().String()
}

// SaveRefreshToken имплементирующем интерфейс
func (s *Service) SaveRefreshToken(ctx context.Context, userID, refreshToken string) error {
	err := s.repo.SaveRefreshToken(ctx, userID, refreshToken)
	if err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken - имплементирующем интерфейс
func (s *Service) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	// Пробуем получить из Redis
	refreshToken, err := s.repo.GetRefreshToken(ctx, userID)
	if err == nil && refreshToken != "" {
		return refreshToken, nil
	}

	// Если нет в Redis пробуем получить из UsersService
	token, err := s.client.GetRefreshToken(ctx, UsersClient.TokenRequest{ID: userID})
	if err != nil || token == "" {
		return "", fmt.Errorf("get refresh token: not found in Redis and UsersService: %w", err)
	}

	// Кэшируем обратно в Redis
	_ = s.repo.SaveRefreshToken(ctx, userID, token)

	return token, nil
}

// GenerateNewRefreshToken - создаем новый refresh токен
func (s *Service) GenerateNewRefreshToken() string {
	return uuid.NewString()
}

// GenerateAccessToken - создаем новый access(jwt) токен
func (s *Service) GenerateAccessToken(userID, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(accessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	s.log.Info("Access token generated successfully", zap.String("userID", userID))

	return tokenString, nil
}

// UpdateRefreshToken - обновляем устаревший refresh токен пользователя
func (s *Service) UpdateRefreshToken(ctx context.Context, userID string) (string, error) {
	newRefreshToken := s.GenerateNewRefreshToken()
	err := s.repo.SaveRefreshToken(ctx, userID, newRefreshToken)
	if err != nil {
		return "", fmt.Errorf("save refresh token: %w", err)
	}

	s.log.Info("Update refresh token successfully")

	return newRefreshToken, nil
}

// ExtractUserIDFromToken взять userID из заголовка токена
func (s *Service) ExtractUserIDFromToken(tokenStr, secret string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		//проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("ExtractUserIDFromToken: invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("ExtractUserIDFromToken: cannot parse claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("ExtractUserIDFromToken: userID not found in token claims")
	}

	return userID, nil
}

// RegisterUser - регистрируем нового пользователя и посылаем данные в Users Service
func (s *Service) RegisterUser(ctx context.Context, reg entity.RegisterRequest) (string, string, string, error) {
	userID := s.GenerateUserID()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", "", fmt.Errorf("RegisterUser: failed to hash password: %w", err)
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
		return "", "", "", fmt.Errorf("RegisterUser: failed to register user: %w", err)
	}

	refreshToken := s.GenerateNewRefreshToken()
	err = s.repo.SaveRefreshToken(ctx, userID, refreshToken)
	if err != nil {
		return "", "", "", fmt.Errorf("RegisterUser: failed to save refresh token: %w", err)
	}

	accessToken, err := s.GenerateAccessToken(userID, s.cfg.JWTSecret)
	if err != nil {
		return "", "", "", fmt.Errorf("RegisterUser: failed to generate access token: %w", err)
	}

	return userID, accessToken, refreshToken, nil
}

// LoginAccount - входим в аккаунт и отдаем новые токены
func (s *Service) LoginAccount(ctx context.Context, login, password string) (string, string, string, error) {
	userID, err := s.client.CompareAuthData(ctx, UsersClient.AuthRequest{
		Login:    login,
		Password: password,
	})

	if err != nil {
		return "", "", "", fmt.Errorf("LoginAccount: failed to compare auth data: %w", err)
	}

	//обновляем токен в redis
	newRefreshToken, err := s.UpdateRefreshToken(ctx, userID)
	if err != nil {
		return "", "", "", fmt.Errorf("LoginAccount: failed to update refresh token: %w", err)
	}

	////обновляем токен в redis
	newAccessToken, err := s.GenerateAccessToken(userID, s.cfg.JWTSecret)
	if err != nil {
		return "", "", "", fmt.Errorf("LoginAccount: failed to generate access token: %w", err)
	}

	// обновляем токен через клиента UpdateRefreshToken
	var token UsersClient.UpdateRefreshTokenRequest

	token.ID = userID
	token.RefreshToken = newRefreshToken

	err = s.client.UpdateRefreshToken(ctx, token)

	return userID, newAccessToken, newRefreshToken, nil
}

// RefreshTokens - Отдаем новые токены
func (s *Service) RefreshTokens(ctx context.Context, refreshToken, authHeader string) (string, string, error) {
	// Checking the Authorization Header Format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", "", fmt.Errorf("RefreshTokens: invalid authorization header")
	}

	tokenString := parts[1]

	// Get userID from access token
	userID, err := s.ExtractUserIDFromToken(tokenString, s.cfg.JWTSecret)
	if err != nil {
		return "", "", fmt.Errorf("RefreshTokens: invalid access token: %w", err)
	}

	// Получаем сохраненный refresh token
	storedRefreshToken, err := s.GetRefreshToken(ctx, userID)
	if err != nil || storedRefreshToken == "" {
		// Если не найден — генерируем новый, сохраняем в Redis и обновляем в UsersService
		newRefreshToken := s.GenerateNewRefreshToken()
		_ = s.SaveRefreshToken(ctx, userID, newRefreshToken)
		_ = s.client.UpdateRefreshToken(ctx, UsersClient.UpdateRefreshTokenRequest{
			ID:           userID,
			RefreshToken: newRefreshToken,
		})
		storedRefreshToken = newRefreshToken
	}

	// Сравниваем токены
	if storedRefreshToken != refreshToken {
		return "", "", fmt.Errorf("RefreshTokens: refresh token mismatch")
	}

	// Generate a new refresh token
	newRefreshToken := s.GenerateNewRefreshToken()
	if err := s.SaveRefreshToken(ctx, userID, newRefreshToken); err != nil {
		return "", "", fmt.Errorf("RefreshTokens: failed to save refresh token: %w", err)
	}
	_ = s.client.UpdateRefreshToken(ctx, UsersClient.UpdateRefreshTokenRequest{
		ID:           userID,
		RefreshToken: newRefreshToken,
	})

	// Generate a new access token
	newAccessToken, err := s.GenerateAccessToken(userID, s.cfg.JWTSecret)
	if err != nil {
		return "", "", fmt.Errorf("RefreshTokens: failed to generate access token: %w", err)
	}

	return newRefreshToken, newAccessToken, nil
}
