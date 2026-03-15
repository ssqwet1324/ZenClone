package usecase

import (
	"UsersService/internal/entity"
	pkgMetrics "UsersService/internal/pkg/prometheus"
	"context"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const nameTracer = "usecase_users"

// UsecaseObs - обёртка над Usecase с метриками и трассировкой
type UsecaseObs struct {
	*Usecase
	metrics *pkgMetrics.Metrics
}

// NewObs - конструктор обёртки с метриками
func NewObs(uc *Usecase) *UsecaseObs {
	return &UsecaseObs{
		Usecase: uc,
		metrics: pkgMetrics.NewMetrics(nameTracer),
	}
}

// observe - сбор метрик и трассировки для каждого метода
func (uc *UsecaseObs) observe(ctx context.Context, method string, fn func(context.Context) error) error {
	ctx, span := otel.Tracer(nameTracer).Start(ctx, method)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start).Seconds()

	if err != nil {
		uc.metrics.HitError(method)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to uc.Usecase."+method)
	} else {
		uc.metrics.HitSuccess(method)
	}

	uc.metrics.HitDuration(method, duration)

	return err
}

// AddUser - метрики + трассировка для добавления пользователя
func (uc *UsecaseObs) AddUser(ctx context.Context, addUserInfo entity.AddUserRequest) error {
	const method = "AddUser"
	return uc.observe(ctx, method, func(ctx context.Context) error {
		return uc.Usecase.AddUser(ctx, addUserInfo)
	})
}

// CompareAuthData - метрики + трассировка для сравнения данных авторизации
func (uc *UsecaseObs) CompareAuthData(ctx context.Context, users entity.AuthRequest) (*entity.CompareDataResponse, error) {
	const method = "CompareAuthData"
	var resp *entity.CompareDataResponse

	err := uc.observe(ctx, method, func(ctx context.Context) error {
		r, err := uc.Usecase.CompareAuthData(ctx, users)
		resp = r
		return err
	})

	return resp, err
}

// UpdateUserProfile - метрики + трассировка для обновления профиля пользователя
func (uc *UsecaseObs) UpdateUserProfile(ctx context.Context, id uuid.UUID, updateProfileInfo entity.UpdateUserProfileInfoRequest) (*entity.UpdateUserProfileInfoResponse, error) {
	const method = "UpdateUserProfile"
	var resp *entity.UpdateUserProfileInfoResponse

	err := uc.observe(ctx, method, func(ctx context.Context) error {
		r, err := uc.Usecase.UpdateUserProfile(ctx, id, updateProfileInfo)
		resp = r
		return err
	})

	return resp, err
}

// GetUserProfileByID - метрики + трассировка для получения профиля по ID
func (uc *UsecaseObs) GetUserProfileByID(ctx context.Context, yourUserID uuid.UUID, otherUserID uuid.UUID) (*entity.ProfileUserInfoResponse, error) {
	const method = "GetUserProfileByID"
	var resp *entity.ProfileUserInfoResponse

	err := uc.observe(ctx, method, func(ctx context.Context) error {
		r, err := uc.Usecase.GetUserProfileByID(ctx, yourUserID, otherUserID)
		resp = r
		return err
	})

	return resp, err
}

// SubscribeToUser - метрики + трассировка для подписки на пользователя
func (uc *UsecaseObs) SubscribeToUser(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) error {
	const method = "SubscribeToUser"
	return uc.observe(ctx, method, func(ctx context.Context) error {
		return uc.Usecase.SubscribeToUser(ctx, followerID, followingID)
	})
}

// UnsubscribeFromUser - метрики + трассировка для отписки от пользователя
func (uc *UsecaseObs) UnsubscribeFromUser(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) error {
	const method = "UnsubscribeFromUser"
	return uc.observe(ctx, method, func(ctx context.Context) error {
		return uc.Usecase.UnsubscribeFromUser(ctx, followerID, followingID)
	})
}

// GetSubsUser - метрики + трассировка для получения подписок пользователя
func (uc *UsecaseObs) GetSubsUser(ctx context.Context, userID uuid.UUID) (*entity.SubsList, error) {
	const method = "GetSubsUser"
	var resp *entity.SubsList

	err := uc.observe(ctx, method, func(ctx context.Context) error {
		r, err := uc.Usecase.GetSubsUser(ctx, userID)
		resp = r
		return err
	})

	return resp, err
}

// UploadAvatar - метрики + трассировка для загрузки аватара
func (uc *UsecaseObs) UploadAvatar(ctx context.Context, userID uuid.UUID, avatarInfo entity.AvatarRequest) error {
	const method = "UploadAvatar"
	return uc.observe(ctx, method, func(ctx context.Context) error {
		return uc.Usecase.UploadAvatar(ctx, userID, avatarInfo)
	})
}

// GlobalSearchPeople - метрики + трассировка для глобального поиска людей
func (uc *UsecaseObs) GlobalSearchPeople(ctx context.Context, firstName, lastName string) (*entity.PersonDateList, error) {
	const method = "GlobalSearchPeople"
	var resp *entity.PersonDateList

	err := uc.observe(ctx, method, func(ctx context.Context) error {
		r, err := uc.Usecase.GlobalSearchPeople(ctx, firstName, lastName)
		resp = r
		return err
	})

	return resp, err
}
