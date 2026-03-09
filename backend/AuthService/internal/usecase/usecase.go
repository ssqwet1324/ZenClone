package usecase

import (
	"AuthService/internal/client/usersclient"
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

// RepositoryProvider - интерфейс repository
type RepositoryProvider interface {
	SaveRefreshToken(ctx context.Context, userID, refreshToken string) error
	GetRefreshToken(ctx context.Context, userID string) (string, error)
}

// UseCaseInterface - интерфейс для метрик и тресинга
type UseCaseInterface interface {
	RegisterUser(ctx context.Context, reg entity.RegisterRequest) (*entity.RegisterResponse, error)
	LoginAccount(ctx context.Context, login, password string) (*entity.LoginResponse, error)
	RefreshTokens(ctx context.Context, refreshToken, authHeader string) (*entity.RefreshResponse, error)
}

// UseCase - бизнес логика
type UseCase struct {
	repo   RepositoryProvider
	log    *zap.Logger
	client usersclient.ClientProvider
	cfg    *config.Config
}

// New - конструктор
func New(repo RepositoryProvider, log *zap.Logger, client usersclient.ClientProvider, cfg *config.Config) UseCaseInterface {
	usecase := &UseCase{
		repo:   repo,
		log:    log,
		client: client,
		cfg:    cfg,
	}

	return NewObs(usecase)
}

// generateUserID - генерируем уникальный id пользователя
func generateUserID() string {
	return uuid.New().String()
}

// saveRefreshToken имплементирующем интерфейс
func (s *UseCase) saveRefreshToken(ctx context.Context, userID, refreshToken string) error {
	err := s.repo.SaveRefreshToken(ctx, userID, refreshToken)
	if err != nil {
		s.log.Error("Failed to save refresh token",
			zap.String("userID", userID),
		)
		return entity.ErrSaveRefreshToken
	}

	return nil
}

// getRefreshToken - имплементирующем интерфейс
func (s *UseCase) getRefreshToken(ctx context.Context, userID string) (string, error) {
	// Пробуем получить из Redis
	refreshToken, err := s.repo.GetRefreshToken(ctx, userID)
	if err == nil && refreshToken != "" {
		return refreshToken, nil
	}

	// Если нет в Redis пробуем получить из UsersService
	token, err := s.client.GetRefreshToken(ctx, usersclient.TokenRequest{ID: userID})
	if err != nil || token == "" {
		s.log.Error("Failed to get refresh token from Redis and UsersService",
			zap.String("userID", userID),
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

// generateAccessToken - создаем новый access(jwt) токен
func (s *UseCase) generateAccessToken(userID, secretKey string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(accessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		s.log.Error("generateAccessToken: Failed to sign token",
			zap.String("userID", userID),
		)
		return "", entity.ErrSignToken
	}

	return tokenString, nil
}

// updateRefreshToken - обновляем устаревший refresh токен пользователя
func (s *UseCase) updateRefreshToken(ctx context.Context, userID string) (string, error) {
	newRefreshToken := generateNewRefreshToken()
	err := s.repo.SaveRefreshToken(ctx, userID, newRefreshToken)
	if err != nil {
		s.log.Error("updateRefreshToken: Failed to save refresh token in updateRefreshToken",
			zap.String("userID", userID),
		)
		return "", entity.ErrSaveRefreshToken
	}

	return newRefreshToken, nil
}

// extractUserIDFromToken взять userID из заголовка токена
func (s *UseCase) extractUserIDFromToken(tokenStr, secret string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		//проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			s.log.Error("extractUserIDFromToken: Unexpected signing method")
			return nil, entity.ErrUnexpectedSigningMethod
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		s.log.Error("extractUserIDFromToken: Invalid token",
			zap.Error(err),
		)
		return "", entity.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		s.log.Error("extractUserIDFromToken: Cannot parse claims")
		return "", entity.ErrCannotParseClaims
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		s.log.Error("extractUserIDFromToken: UserID not found in token claims")
		return "", entity.ErrUserIDNotFound
	}

	return userID, nil
}

// RegisterUser - регистрируем нового пользователя и посылаем данные в Users UseCase
func (s *UseCase) RegisterUser(ctx context.Context, reg entity.RegisterRequest) (*entity.RegisterResponse, error) {
	var resp entity.RegisterResponse
	userID := generateUserID()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("RegisterUser: Failed to hash password",
			zap.String("login", reg.Login),
		)
		return nil, entity.ErrHashPassword
	}

	err = s.client.AddUser(ctx, usersclient.RegisterRequest{
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
		return nil, err
	}

	refreshToken := generateNewRefreshToken()
	err = s.repo.SaveRefreshToken(ctx, userID, refreshToken)
	if err != nil {
		s.log.Error("RegisterUser: Failed to save refresh token in RegisterUser",
			zap.String("userID", userID),
		)
		return nil, entity.ErrSaveRefreshToken
	}

	accessToken, err := s.generateAccessToken(userID, s.cfg.JWTSecret)
	if err != nil {
		s.log.Error("RegisterUser: Failed to generate access token in RegisterUser",
			zap.String("userID", userID),
		)
		return nil, entity.ErrGenerateAccessToken
	}

	resp.ID = userID
	resp.AccessToken = accessToken
	resp.RefreshToken = refreshToken

	return &resp, nil
}

// LoginAccount - входим в аккаунт и отдаем новые токены
func (s *UseCase) LoginAccount(ctx context.Context, login, password string) (*entity.LoginResponse, error) {
	response, err := s.client.CompareAuthData(ctx, usersclient.AuthRequest{
		Login:    login,
		Password: password,
	})

	if err != nil {
		s.log.Error("LoginAccount: Failed to compare auth data",
			zap.String("login", login),
		)
		return nil, err
	}

	//обновляем токен в redis
	newRefreshToken, err := s.updateRefreshToken(ctx, response.ID)
	if err != nil {
		s.log.Error("LoginAccount: Failed to update refresh token in LoginAccount",
			zap.String("userID", response.ID),
			zap.Error(err),
		)
		return nil, entity.ErrUpdateRefreshToken
	}

	//обновляем токен в redis
	newAccessToken, err := s.generateAccessToken(response.ID, s.cfg.JWTSecret)
	if err != nil {
		s.log.Error("LoginAccount: Failed to generate access token in LoginAccount",
			zap.String("userID", response.ID),
			zap.Error(err),
		)
		return nil, entity.ErrGenerateAccessToken
	}

	// обновляем токен через клиента updateRefreshToken
	var token usersclient.UpdateRefreshTokenRequest

	token.ID = response.ID
	token.RefreshToken = newRefreshToken

	err = s.client.UpdateRefreshToken(ctx, token)
	if err != nil {
		s.log.Error("LoginAccount: Failed to update refresh token in LoginAccount",
			zap.String("userID", response.ID),
			zap.Error(err))
		return nil, entity.ErrUpdateRefreshToken
	}

	var LoginResponse entity.LoginResponse
	LoginResponse.ID = response.ID
	LoginResponse.Username = response.Username
	LoginResponse.AccessToken = newAccessToken
	LoginResponse.RefreshToken = newRefreshToken

	return &LoginResponse, nil
}

// RefreshTokens - Отдаем новые токены
func (s *UseCase) RefreshTokens(ctx context.Context, refreshToken, authHeader string) (*entity.RefreshResponse, error) {
	var resp entity.RefreshResponse

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		s.log.Error("RefreshTokens: Invalid authorization header")
		return nil, entity.ErrInvalidAuthHeader
	}

	tokenString := parts[1]

	// Get userID from access token
	userID, err := s.extractUserIDFromToken(tokenString, s.cfg.JWTSecret)
	if err != nil {
		s.log.Error("RefreshTokens: Invalid access token in RefreshTokens")
		return nil, entity.ErrInvalidToken
	}

	// Получаем сохраненный refresh token
	storedRefreshToken, err := s.getRefreshToken(ctx, userID)
	if err != nil || storedRefreshToken == "" {
		// Если не найден — генерируем новый, сохраняем в Redis и обновляем в UsersService
		newRefreshToken := generateNewRefreshToken()
		_ = s.saveRefreshToken(ctx, userID, newRefreshToken)
		_ = s.client.UpdateRefreshToken(ctx, usersclient.UpdateRefreshTokenRequest{
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
		return nil, entity.ErrRefreshTokenMismatch
	}

	// Generate a new refresh token
	newRefreshToken := generateNewRefreshToken()
	if err := s.saveRefreshToken(ctx, userID, newRefreshToken); err != nil {
		s.log.Error("RefreshTokens: Failed to save refresh token in RefreshTokens",
			zap.String("userID", userID),
		)
		return nil, entity.ErrSaveRefreshToken
	}
	_ = s.client.UpdateRefreshToken(ctx, usersclient.UpdateRefreshTokenRequest{
		ID:           userID,
		RefreshToken: newRefreshToken,
	})

	// Generate a new access token
	newAccessToken, err := s.generateAccessToken(userID, s.cfg.JWTSecret)
	if err != nil {
		s.log.Error("RefreshTokens: Failed to generate access token in RefreshTokens",
			zap.String("userID", userID),
		)
		return nil, entity.ErrGenerateAccessToken
	}

	resp.RefreshToken = newRefreshToken
	resp.AccessToken = newAccessToken

	return &resp, nil
}
