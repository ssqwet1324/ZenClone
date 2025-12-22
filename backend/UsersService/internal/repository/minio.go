package repository

import (
	"UsersService/internal/entity"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// UploadAvatar - загружаем фото и сохраняем его имя в бд
func (repo *PostgresRepository) UploadAvatar(ctx context.Context, userID uuid.UUID, bucketName string, avatarInfo entity.AvatarRequest) error {
	// timestamp, чтобы ссылка менялась при каждой загрузке
	objectName := fmt.Sprintf("%s/avatar_%d.jpg", userID.String(), time.Now().Unix())

	_, err := repo.client.PutObject(ctx, bucketName, objectName, avatarInfo.Reader, avatarInfo.Size, minio.PutObjectOptions{
		ContentType: "image/jpg",
	})
	if err != nil {
		return fmt.Errorf("UploadAvatar: error uploading avatar: %w", err)
	}

	// Сохраняем путь к аватару в БД
	_, err = repo.db.Exec(ctx, `UPDATE users SET avatar_url = $1 WHERE id = $2`, objectName, userID)
	if err != nil {
		return fmt.Errorf("UploadAvatar: error saving avatar_url in db: %w", err)
	}

	return nil
}

// GetAvatarURL - получаем url аватарки по его имени
func (repo *PostgresRepository) GetAvatarURL(ctx context.Context, bucketName string, userID uuid.UUID) (string, error) {
	var objectName string
	err := repo.db.QueryRow(ctx, `SELECT avatar_url FROM users WHERE id = $1`, userID).Scan(&objectName)
	if err != nil {
		return "", fmt.Errorf("GetAvatarURL: error fetching avatar_url from db: %w", err)
	}

	publicEndpoint := repo.config.MinIoPublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = repo.config.MinioEndpoint
	}

	if objectName == "default" {
		// Возвращаем дефолтную аватарку
		avatarURL := fmt.Sprintf("%s/%s/%s", publicEndpoint, bucketName, defaultAvatar)
		return avatarURL, nil
	}

	avatarURL := fmt.Sprintf("%s/%s/%s", publicEndpoint, bucketName, objectName)
	return avatarURL, nil
}
