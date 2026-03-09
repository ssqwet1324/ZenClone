package usecase

import (
	"AuthService/internal/entity"
	pkgmetrics "AuthService/internal/pkg/prometheus"
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const nameTracer = "usecase"

type UseCaseObs struct {
	*UseCase
	metrics *pkgmetrics.Metrics
}

func NewObs(uc *UseCase) *UseCaseObs {
	return &UseCaseObs{
		UseCase: uc,
		metrics: pkgmetrics.NewMetrics(nameTracer),
	}
}

func (uc *UseCaseObs) observe(ctx context.Context, method string, fn func(context.Context) error) error {
	ctx, span := otel.Tracer(nameTracer).Start(ctx, method)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start).Seconds()

	if err != nil {
		uc.metrics.HitError(method)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to uc.UseCase."+method)
	} else {
		uc.metrics.HitSuccess(method)
	}

	uc.metrics.HitDuration(method, duration)

	return err
}

// RegisterUser - метрики + трассировка
func (uc *UseCaseObs) RegisterUser(ctx context.Context, reg entity.RegisterRequest) (*entity.RegisterResponse, error) {
	const method = "RegisterUser"
	var resp *entity.RegisterResponse

	err := uc.observe(ctx, method, func(ctx context.Context) error {
		r, err := uc.UseCase.RegisterUser(ctx, reg)
		resp = r
		return err
	})

	return resp, err
}

// LoginAccount - метрики + трасировка
func (uc *UseCaseObs) LoginAccount(ctx context.Context, login, password string) (*entity.LoginResponse, error) {
	const method = "LoginAccount"
	var resp *entity.LoginResponse

	err := uc.observe(ctx, method, func(ctx context.Context) error {
		r, err := uc.UseCase.LoginAccount(ctx, login, password)
		resp = r
		return err
	})

	return resp, err
}

// RefreshTokens - метрики + трасировка
func (uc *UseCaseObs) RefreshTokens(ctx context.Context, refreshToken, authHeader string) (*entity.RefreshResponse, error) {
	const method = "RefreshTokens"
	var resp *entity.RefreshResponse

	err := uc.observe(ctx, method, func(ctx context.Context) error {
		r, err := uc.UseCase.RefreshTokens(ctx, refreshToken, authHeader)
		resp = r
		return err
	})

	return resp, err
}
