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

func (m *MinioClientMock) ListObjects(ctx context.Context, bucketName string, opts minio.ListObjectsOptions) <-chan minio.ObjectInfo {
	args := m.Called(ctx, bucketName, opts)
	if ch, ok := args.Get(0).(<-chan minio.ObjectInfo); ok {
		return ch
	}
	ch := make(chan minio.ObjectInfo)
	close(ch)
	return ch
}

func (m *MinioClientMock) RemoveObjects(ctx context.Context, bucketName string, objectsCh <-chan minio.ObjectInfo, opts minio.RemoveObjectsOptions) <-chan minio.RemoveObjectError {
	args := m.Called(ctx, bucketName, objectsCh, opts)
	if ch, ok := args.Get(0).(<-chan minio.RemoveObjectError); ok {
		return ch
	}
	ch := make(chan minio.RemoveObjectError)
	close(ch)
	return ch
}

func (m *MinioClientMock) CopyObject(ctx context.Context, dst minio.CopyDestOptions, src minio.CopySrcOptions) (minio.UploadInfo, error) {
	args := m.Called(ctx, dst, src)
	if info, ok := args.Get(0).(minio.UploadInfo); ok {
		return info, args.Error(1)
	}
	return minio.UploadInfo{}, args.Error(1)
}
