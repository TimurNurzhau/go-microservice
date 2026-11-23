package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func InitMinIO() (*minio.Client, error) {
	// Читаем настройки из переменных окружения (для Docker-сети)
	// Если переменных нет, используем публичный play.min.io
	endpoint := getEnv("MINIO_ENDPOINT", "play.min.io")
	accessKeyID := getEnv("MINIO_ACCESS_KEY", "Q3AM3UQ867SPQQA43P2F")
	secretAccessKey := getEnv("MINIO_SECRET_KEY", "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG")
	useSSL := getEnv("MINIO_USE_SSL", "true") == "true"

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания клиента MinIO: %w", err)
	}

	bucketName := "users-avatars"
	ctx := context.Background()

	// Проверяем, существует ли бакет
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки бакета: %w", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("ошибка создания бакета: %w", err)
		}
		log.Printf("MinIO: бакет %s успешно создан\n", bucketName)
	} else {
		log.Printf("MinIO: бакет %s уже существует\n", bucketName)
	}

	// Загружаем тестовый файл (исправлено: передаём strings.NewReader вместо nil)
	objectName := "test-connection.txt"
	content := "MinIO integration successfully initialized!"
	_, err = minioClient.PutObject(ctx, bucketName, objectName,
		strings.NewReader(content), int64(len(content)), minio.PutObjectOptions{ContentType: "text/plain"})
	if err != nil {
		log.Printf("MinIO: предупреждение при загрузке тестового файла: %v\n", err)
	} else {
		fmt.Println("MinIO: тестовое подключение успешно")
	}

	return minioClient, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}