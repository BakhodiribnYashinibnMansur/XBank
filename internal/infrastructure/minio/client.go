// Package minio provides S3-compatible file storage via MinIO.
package minio

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// Config holds MinIO connection parameters.
type Config struct {
	Endpoint  string // e.g. "localhost:9000"
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string // default bucket
}

// Client wraps the MinIO SDK.
type Client struct {
	mc     *minio.Client
	bucket string
}

// NewClient creates and validates a MinIO client.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio connect: %w", err)
	}

	// Ensure bucket exists
	exists, err := mc.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("minio bucket check: %w", err)
	}
	if !exists {
		if err := mc.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio create bucket: %w", err)
		}
		logger.Log.Info("MinIO bucket created", zap.String("bucket", cfg.Bucket))
	}

	logger.Log.Info("MinIO connected", zap.String("endpoint", cfg.Endpoint), zap.String("bucket", cfg.Bucket))
	return &Client{mc: mc, bucket: cfg.Bucket}, nil
}

// Upload stores a file and returns its path.
func (c *Client) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := c.mc.PutObject(ctx, c.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("minio upload: %w", err)
	}
	return fmt.Sprintf("%s/%s", c.bucket, objectName), nil
}

// Download retrieves a file.
func (c *Client) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	obj, err := c.mc.GetObject(ctx, c.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("minio download: %w", err)
	}
	return obj, nil
}

// Delete removes a file.
func (c *Client) Delete(ctx context.Context, objectName string) error {
	return c.mc.RemoveObject(ctx, c.bucket, objectName, minio.RemoveObjectOptions{})
}

// PresignedURL generates a temporary download URL.
func (c *Client) PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := c.mc.PresignedGetObject(ctx, c.bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("minio presign: %w", err)
	}
	return url.String(), nil
}
