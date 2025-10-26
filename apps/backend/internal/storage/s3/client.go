package s3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioClient описывает взаимодействие с MinIO SDK.
type MinioClient interface {
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
	SetBucketPolicy(ctx context.Context, bucketName, policy string) error
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	EndpointURL() *url.URL
	RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error
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

		objectName := filepath.Join(remotePath, relPath)

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
		remotePath = filepath.Join(remotePath, "index.html")
	}

	return fmt.Sprintf("%s://%s/%s/%s", scheme, c.minio.EndpointURL().Host, c.bucket, remotePath)
}

// Delete удаляет объект
func (c *Client) Delete(ctx context.Context, remotePath string) error {
	return c.minio.RemoveObject(ctx, c.bucket, remotePath, minio.RemoveObjectOptions{})
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
