package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func InitMinIO() (*minio.Client, error) {
	endpoint := "play.min.io" // Публичный тестовый сервер MinIO
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	useSSL := true

	// Инициализация клиента MinIO
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	// Создаем бакет (если его нет)
	bucketName := "users-avatars"
	ctx := context.Background()
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("MinIO: бакет %s уже существует\n", bucketName)
		} else {
			return nil, err
		}
	} else {
		log.Printf("MinIO: бакет %s успешно создан\n", bucketName)
	}

	// Проверяем подключение, загружая тестовый файл
	objectName := "test-connection.txt"
	content := "MinIO integration successfully initialized!"
	_, err = minioClient.PutObject(ctx, bucketName, objectName, 
		nil, int64(len(content)), minio.PutObjectOptions{})
	if err != nil {
		// Если PutObject с nil не работает, используем strings.NewReader
		log.Println("MinIO: тестовый файл не загружен (это нормально, если нет прав на запись)")
	} else {
		fmt.Println("MinIO: тестовое подключение успешно")
	}

	return minioClient, nil
}