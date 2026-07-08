package minio

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

type Client struct {
	mc        *minio.Client
	bucket    string
	publicURL string
}

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	PublicURL string
	UseSSL    bool
}

func New(cfg Config) (*Client, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio.New: %w", err)
	}

	return &Client{mc: mc, bucket: cfg.Bucket, publicURL: strings.TrimRight(cfg.PublicURL, "/")}, nil
}

func (c *Client) Upload(ctx context.Context, folder string, file io.Reader, contentLength int64, contentType string) (string, error) {
	ctx, span := otel.StartSpan(ctx, "minio.Upload")
	defer span.End()

	ext := extractExt(contentType)
	objectName := fmt.Sprintf("%s/%s%s", folder, uuid.NewString()[:16], ext)

	_, err := c.mc.PutObject(ctx, c.bucket, objectName, file, contentLength, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		logger.Error("minio.Upload: PutObject failed", "error", err.Error())
		return "", fmt.Errorf("minio.Upload: %w", err)
	}

	return path.Join(c.publicURL, c.bucket, objectName), nil
}

func extractExt(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "image/svg+xml":
		return ".svg"
	default:
		return ".bin"
	}
}
