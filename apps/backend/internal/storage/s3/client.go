package s3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/landly/backend/internal/logger"
	domain "github.com/landly/backend/internal/models"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// FileManifest представляет manifest файла для incremental upload
type FileManifest struct {
	Path     string    `json:"path"`
	Hash     string    `json:"hash"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

// ReleaseManifest представляет manifest релиза
type ReleaseManifest struct {
	ReleaseID uuid.UUID      `json:"release_id"`
	Files     []FileManifest `json:"files"`
	CreatedAt time.Time      `json:"created_at"`
}

// MinioClient описывает взаимодействие с MinIO SDK.
type MinioClient interface {
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
	SetBucketPolicy(ctx context.Context, bucketName, policy string) error
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error)
	CopyObject(ctx context.Context, dst minio.CopyDestOptions, src minio.CopySrcOptions) (minio.UploadInfo, error)
	EndpointURL() *url.URL
	RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
	ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo
	RemoveObjects(ctx context.Context, bucketName string, objectsCh <-chan minio.ObjectInfo, opts minio.RemoveObjectsOptions) <-chan minio.RemoveObjectError
}

// Config конфигурация S3 клиента
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
	CDNBase         string
}

// Client S3-совместимый клиент для хранения файлов
// PLUGGABLE: работает с любым S3-совместимым хранилищем
type Client struct {
	minio   MinioClient
	bucket  string
	cdnBase string
}

// Option конфигурирует S3 клиент.
type Option func(*Client)

// WithMinioClient позволяет подменить реализацию (используется в тестах).
func WithMinioClient(m MinioClient) Option {
	return func(c *Client) {
		c.minio = m
	}
}

// NewClient создаёт новый S3 клиент
func NewClient(cfg Config, opts ...Option) (*Client, error) {
	client := &Client{bucket: cfg.BucketName, cdnBase: cfg.CDNBase}

	for _, opt := range opts {
		opt(client)
	}

	if client.minio == nil {
		minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
			Secure: cfg.UseSSL,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create minio client: %w", err)
		}
		client.minio = minioClient
	}

	ctx := context.Background()
	exists, err := client.minio.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		if err := client.minio.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}

		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{"Effect": "Allow", "Principal": {"AWS": ["*"]}, "Action": ["s3:GetObject"], "Resource": ["arn:aws:s3:::%s/*"]}]}`, cfg.BucketName)

		if err := client.minio.SetBucketPolicy(ctx, cfg.BucketName, policy); err != nil {
			return nil, fmt.Errorf("failed to set bucket policy: %w", err)
		}
	}

	return client, nil
}

// Upload загружает директорию в S3
func (c *Client) Upload(ctx context.Context, localPath, remotePath string) error {
	return filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return err
		}

		// Конвертируем путь в формат S3 (прямые слэши)
		relPath = strings.ReplaceAll(relPath, "\\", "/")
		remotePath = strings.TrimSuffix(remotePath, "/")
		objectName := remotePath + "/" + relPath

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		contentType := getContentType(path)

		_, err = c.minio.PutObject(ctx, c.bucket, objectName, file, info.Size(), minio.PutObjectOptions{ContentType: contentType})
		return err
	})
}

// UploadFile загружает один файл
func (c *Client) UploadFile(ctx context.Context, reader io.Reader, remotePath string, size int64) error {
	contentType := getContentType(remotePath)
	_, err := c.minio.PutObject(ctx, c.bucket, remotePath, reader, size, minio.PutObjectOptions{ContentType: contentType})
	return err
}

// GetPublicURL возвращает публичный URL файла
func (c *Client) GetPublicURL(remotePath string) string {
	if c.cdnBase != "" {
		return fmt.Sprintf("%s/%s", c.cdnBase, remotePath)
	}

	scheme := "http"
	if c.minio.EndpointURL().Scheme == "https" {
		scheme = "https"
	}

	if !strings.HasSuffix(remotePath, "/index.html") {
		remotePath = strings.TrimSuffix(remotePath, "/") + "/index.html"
	}

	return fmt.Sprintf("%s://%s/%s/%s", scheme, c.minio.EndpointURL().Host, c.bucket, remotePath)
}

// HasCDN проверяет, настроен ли CDN
func (c *Client) HasCDN() bool {
	return c.cdnBase != ""
}

// GetObject возвращает содержимое объекта и его content-type
func (c *Client) GetObject(ctx context.Context, remotePath string) (io.ReadCloser, string, error) {
	log := logger.WithContext(ctx).With(
		zap.String("bucket", c.bucket),
		zap.String("remote_path", remotePath),
	)

	log.Debug("GetObject called",
		zap.String("bucket", c.bucket),
		zap.String("remote_path", remotePath))

	object, err := c.minio.GetObject(ctx, c.bucket, remotePath, minio.GetObjectOptions{})
	if err != nil {
		if resp := minio.ToErrorResponse(err); resp.Code == "NoSuchKey" {
			log.Debug("object not found in S3",
				zap.String("remote_path", remotePath),
				zap.String("error_code", resp.Code),
				zap.String("error_message", resp.Message))
			return nil, "", domain.ErrNotFound
		}
		log.Error("error getting object from S3",
			zap.String("remote_path", remotePath),
			zap.String("error_type", fmt.Sprintf("%T", err)),
			zap.Error(err))
		return nil, "", err
	}

	info, err := object.Stat()
	if err != nil {
		object.Close()
		if resp := minio.ToErrorResponse(err); resp.Code == "NoSuchKey" {
			log.Debug("object stat failed - not found",
				zap.String("remote_path", remotePath),
				zap.String("error_code", resp.Code))
			return nil, "", domain.ErrNotFound
		}
		log.Error("error getting object stat from S3",
			zap.String("remote_path", remotePath),
			zap.String("error_type", fmt.Sprintf("%T", err)),
			zap.Error(err))
		return nil, "", err
	}

	contentType := info.ContentType
	if contentType == "" {
		contentType = getContentType(remotePath)
		log.Debug("content type not set, using detected type",
			zap.String("detected_content_type", contentType))
	}

	log.Info("object retrieved successfully from S3",
		zap.String("remote_path", remotePath),
		zap.String("content_type", contentType),
		zap.Int64("size", info.Size))

	return object, contentType, nil
}

// Delete удаляет объект
func (c *Client) Delete(ctx context.Context, remotePath string) error {
	return c.minio.RemoveObject(ctx, c.bucket, remotePath, minio.RemoveObjectOptions{})
}

// DeletePrefix удаляет все объекты с указанным префиксом (директорию)
func (c *Client) DeletePrefix(ctx context.Context, prefix string) error {
	// Убеждаемся, что префикс заканчивается на /
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	// Списываем все объекты с данным префиксом
	objectsCh := c.minio.ListObjects(ctx, c.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	// Создаем канал для удаления объектов
	removeObjectsCh := make(chan minio.ObjectInfo, 100)
	go func() {
		defer close(removeObjectsCh)
		for object := range objectsCh {
			if object.Err != nil {
				continue
			}
			removeObjectsCh <- object
		}
	}()

	// Удаляем объекты пакетно
	errorsCh := c.minio.RemoveObjects(ctx, c.bucket, removeObjectsCh, minio.RemoveObjectsOptions{})

	// Проверяем ошибки удаления
	for err := range errorsCh {
		if err.Err != nil {
			return fmt.Errorf("failed to delete object %s: %w", err.ObjectName, err.Err)
		}
	}

	return nil
}

// getContentType определяет MIME-type по расширению файла
func getContentType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	default:
		return "application/octet-stream"
	}
}

// UploadToRelease загружает директорию в release-specific путь
// Структура: sites/{subdomain}/releases/{releaseID}/...
func (c *Client) UploadToRelease(ctx context.Context, localPath, basePath string, releaseID uuid.UUID) error {
	releasePath := c.GetReleasePath(basePath, releaseID)
	return c.Upload(ctx, localPath, releasePath)
}

// CreateManifest создает manifest для всех файлов в директории
func (c *Client) CreateManifest(ctx context.Context, localPath string) (*ReleaseManifest, error) {
	manifest := &ReleaseManifest{
		ReleaseID: uuid.New(),
		Files:     []FileManifest{},
		CreatedAt: time.Now(),
	}

	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return err
		}

		// Вычисляем SHA256 хеш файла
		hash, err := computeFileHash(path)
		if err != nil {
			return fmt.Errorf("failed to compute hash for %s: %w", path, err)
		}

		manifest.Files = append(manifest.Files, FileManifest{
			Path:     relPath,
			Hash:     hash,
			Size:     info.Size(),
			Modified: info.ModTime(),
		})

		return nil
	})

	return manifest, err
}

// UploadIncremental загружает только измененные файлы на основе предыдущего manifest
func (c *Client) UploadIncremental(ctx context.Context, localPath, basePath string, releaseID uuid.UUID, previousManifest *ReleaseManifest) (*ReleaseManifest, error) {
	// Создаем новый manifest для текущей версии
	newManifest, err := c.CreateManifest(ctx, localPath)
	if err != nil {
		return nil, err
	}
	newManifest.ReleaseID = releaseID

	// Создаем map предыдущих файлов для быстрого поиска
	previousFiles := make(map[string]FileManifest)
	if previousManifest != nil {
		for _, file := range previousManifest.Files {
			previousFiles[file.Path] = file
		}
	}

	releasePath := c.GetReleasePath(basePath, releaseID)

	// Загружаем только измененные файлы
	for _, file := range newManifest.Files {
		previousFile, exists := previousFiles[file.Path]
		shouldUpload := !exists || previousFile.Hash != file.Hash

		if shouldUpload {
			fullPath := filepath.Join(localPath, file.Path)
			// Конвертируем путь в формат S3 (прямые слэши)
			filePath := strings.ReplaceAll(file.Path, "\\", "/")
			releasePath = strings.TrimSuffix(releasePath, "/")
			remotePath := releasePath + "/" + filePath

			fileReader, err := os.Open(fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to open file %s: %w", fullPath, err)
			}

			contentType := getContentType(fullPath)
			_, err = c.minio.PutObject(ctx, c.bucket, remotePath, fileReader, file.Size, minio.PutObjectOptions{ContentType: contentType})
			fileReader.Close()

			if err != nil {
				return nil, fmt.Errorf("failed to upload file %s: %w", remotePath, err)
			}
		}
	}

	// Сохраняем manifest в S3
	manifestJSON, err := json.Marshal(newManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}

	releasePath = strings.TrimSuffix(releasePath, "/")
	manifestPath := releasePath + "/.manifest.json"
	manifestReader := strings.NewReader(string(manifestJSON))
	_, err = c.minio.PutObject(ctx, c.bucket, manifestPath, manifestReader, int64(len(manifestJSON)), minio.PutObjectOptions{
		ContentType: "application/json",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload manifest: %w", err)
	}

	return newManifest, nil
}

// SetCurrentRelease атомарно переключает current симлинк на новый release
// В S3 это реализуется через копирование всех файлов из release в current
func (c *Client) SetCurrentRelease(ctx context.Context, basePath string, releaseID uuid.UUID) error {
	releasePath := c.GetReleasePath(basePath, releaseID)
	currentPath := c.GetCurrentReleasePath(basePath)

	// В S3 нет симлинков, поэтому копируем файлы из release в current
	// Это атомарная операция на уровне каждого файла
	objectsCh := c.minio.ListObjects(ctx, c.bucket, minio.ListObjectsOptions{
		Prefix:    releasePath + "/",
		Recursive: true,
	})

	for object := range objectsCh {
		if object.Err != nil {
			continue
		}

		// Пропускаем manifest файл
		if strings.HasSuffix(object.Key, ".manifest.json") {
			continue
		}

		// Вычисляем относительный путь от release
		relPath := strings.TrimPrefix(object.Key, releasePath+"/")
		// Конвертируем путь в формат S3 (прямые слэши)
		relPath = strings.ReplaceAll(relPath, "\\", "/")
		currentPath = strings.TrimSuffix(currentPath, "/")
		currentObjectPath := currentPath + "/" + relPath

		// Копируем объект
		src := minio.CopySrcOptions{
			Bucket: c.bucket,
			Object: object.Key,
		}
		dst := minio.CopyDestOptions{
			Bucket: c.bucket,
			Object: currentObjectPath,
		}

		_, err := c.minio.CopyObject(ctx, dst, src)
		if err != nil {
			return fmt.Errorf("failed to copy object %s to %s: %w", object.Key, currentObjectPath, err)
		}
	}

	return nil
}

// GetCurrentReleasePath возвращает путь к current release
// Использует прямые слэши для S3-совместимых путей
func (c *Client) GetCurrentReleasePath(basePath string) string {
	basePath = strings.TrimSuffix(basePath, "/")
	return basePath + "/current"
}

// GetReleasePath возвращает путь к конкретному release
// Использует прямые слэши для S3-совместимых путей
func (c *Client) GetReleasePath(basePath string, releaseID uuid.UUID) string {
	basePath = strings.TrimSuffix(basePath, "/")
	return basePath + "/releases/" + releaseID.String()
}

// computeFileHash вычисляет SHA256 хеш файла
func computeFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
