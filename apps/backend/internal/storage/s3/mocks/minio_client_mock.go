package mocks

import (
	"context"
	"io"
	"net/url"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/mock"
)

type MinioClientMock struct {
	mock.Mock
}

func (m *MinioClientMock) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	args := m.Called(ctx, bucketName)
	return args.Bool(0), args.Error(1)
}

func (m *MinioClientMock) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	args := m.Called(ctx, bucketName, opts)
	return args.Error(0)
}

func (m *MinioClientMock) SetBucketPolicy(ctx context.Context, bucketName, policy string) error {
	args := m.Called(ctx, bucketName, policy)
	return args.Error(0)
}

func (m *MinioClientMock) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	args := m.Called(ctx, bucketName, objectName, reader, objectSize, opts)
	if info, ok := args.Get(0).(minio.UploadInfo); ok {
		return info, args.Error(1)
	}
	return minio.UploadInfo{}, args.Error(1)
}

func (m *MinioClientMock) GetObject(ctx context.Context, bucketName, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	args := m.Called(ctx, bucketName, objectName, opts)
	object, _ := args.Get(0).(*minio.Object)
	return object, args.Error(1)
}

func (m *MinioClientMock) EndpointURL() *url.URL {
	args := m.Called()
	if endpoint, ok := args.Get(0).(*url.URL); ok {
		return endpoint
	}
	return &url.URL{}
}

func (m *MinioClientMock) RemoveObject(ctx context.Context, bucketName, objectName string, opts minio.RemoveObjectOptions) error {
	args := m.Called(ctx, bucketName, objectName, opts)
	return args.Error(0)
}
