package migrations

import (
	"PostService/internal/repository"
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
	query := `CREATE TABLE IF NOT EXISTS posts (
		post_id UUID PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		content TEXT NOT NULL,
		author_id UUID NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
	);`

	maxRetries := 5
	retryDelay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		_, err := m.repo.DB.Exec(ctx, query)
		if err == nil {
			m.log.Info("Migrations: Table 'posts' created or already exists")

			return nil
		}

		m.log.Warn("Migrations: Failed to init posts table", zap.Error(err))
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("не удалось создать таблицу после %d попыток", maxRetries)
}
