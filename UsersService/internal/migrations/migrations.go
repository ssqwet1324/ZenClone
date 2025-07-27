package migrations

import (
	"UsersService/internal/repository"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Migration struct {
	repo *repository.PostgresRepository
	log  *zap.Logger
}

func New(repo *repository.PostgresRepository, log *zap.Logger) *Migration {
	return &Migration{
		repo: repo,
		log:  log.Named("migrations"),
	}
}

func (m *Migration) InitTables(ctx context.Context) error {
	query := `CREATE TABLE users (
    id UUID PRIMARY KEY,
    login VARCHAR NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,
    refresh_token VARCHAR,
    username VARCHAR NOT NULL UNIQUE, 
    first_name VARCHAR NOT NULL,
    last_name VARCHAR NUll,
    bio TEXT NULL,
    created_at TIMESTAMP DEFAULT now()
)`
	maxRetries := 5
	retryDelay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		_, err := m.repo.DB.Exec(ctx, query)
		if err == nil {
			return nil
		}

		m.log.Warn("Migrations: Failed to init users table", zap.Error(err))
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("неудалось создать таблицу после %d попыток", maxRetries)
}
