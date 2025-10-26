package s3

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/landly/backend/internal/storage/s3/mocks"
)

func TestClient_GetPublicURL_WithCDN(t *testing.T) {
	minioMock := new(mocks.MinioClientMock)
	minioMock.On("BucketExists", context.Background(), "bucket").Return(true, nil)

	client, err := NewClient(Config{BucketName: "bucket"}, WithMinioClient(minioMock))
	require.NoError(t, err)
	client.cdnBase = "https://cdn.example.com"

	assert.Equal(t, "https://cdn.example.com/sites/landing", client.GetPublicURL("sites/landing"))
	minioMock.AssertExpectations(t)
}

func TestClient_GetPublicURL_DefaultAddsIndex(t *testing.T) {
	minioMock := new(mocks.MinioClientMock)
	minioMock.On("BucketExists", context.Background(), "bucket").Return(true, nil)
	minioMock.On("EndpointURL").Return(&url.URL{Scheme: "http", Host: "localhost:9000"})

	client, err := NewClient(Config{BucketName: "bucket"}, WithMinioClient(minioMock))
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:9000/bucket/sites/landing/index.html", client.GetPublicURL("sites/landing"))
	minioMock.AssertExpectations(t)
}

func TestClient_NewClient_BucketExistsError(t *testing.T) {
	minioMock := new(mocks.MinioClientMock)
	minioMock.On("BucketExists", context.Background(), "bucket").Return(false, errors.New("fail"))

	_, err := NewClient(Config{BucketName: "bucket"}, WithMinioClient(minioMock))
	require.Error(t, err)
	minioMock.AssertExpectations(t)
}
