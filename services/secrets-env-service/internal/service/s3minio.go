package service

import (
	"context"
	"io"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioS3Client adapts minio.Client to S3Client
type MinioS3Client struct{ c *minio.Client }

func NewMinioS3Client(endpoint, accessKey, secretKey string, useTLS bool) (*MinioS3Client, error) {
	cli, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(accessKey, secretKey, ""), Secure: useTLS})
	if err != nil { return nil, err }
	return &MinioS3Client{c: cli}, nil
}

func (m *MinioS3Client) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	var keys []string
	for obj := range m.c.ListObjects(ctx, bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
		if obj.Err != nil { return nil, obj.Err }
		keys = append(keys, obj.Key)
	}
	return keys, nil
}

func (m *MinioS3Client) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
    obj, err := m.c.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil { return nil, err }
    defer func(){ _ = obj.Close() }()
	b, err := io.ReadAll(obj)
	if err != nil { return nil, err }
	return b, nil
}
