package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Manager struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	signedURLTTL  time.Duration
}

type ManagerConfig struct {
	AccessKey    string
	SecretKey    string
	Region       string
	Bucket       string
	SignedURLTTL time.Duration
	// Optional: override the S3 endpoint (useful for R2, MinIO, etc.)
	EndpointURL string
}

func NewManager(_ context.Context, cfg ManagerConfig) (*Manager, error) {
	if cfg.SignedURLTTL == 0 {
		cfg.SignedURLTTL = time.Hour
	}
	if cfg.Region == "" {
		cfg.Region = "auto"
	}

	awsCfg := aws.Config{
		Region:      cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
	}

	var clientOptions []func(*s3.Options)
	if cfg.EndpointURL != "" {
		clientOptions = append(clientOptions, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.EndpointURL)
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(awsCfg, clientOptions...)

	return &Manager{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucket:        cfg.Bucket,
		signedURLTTL:  cfg.SignedURLTTL,
	}, nil
}

// PresignGetObject returns a signed URL for downloading an object.
func (m *Manager) PresignGetObject(ctx context.Context, key string) (string, error) {
	req, err := m.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(m.signedURLTTL))
	if err != nil {
		return "", fmt.Errorf("failed to presign GET for key %q: %w", key, err)
	}

	return req.URL, nil
}

// PresignPutObject returns a signed URL the client can use to upload directly to S3.
func (m *Manager) PresignPutObject(ctx context.Context, key string, contentType string, contentLength int64) (string, error) {
	req, err := m.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(m.bucket),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(contentLength),
	}, s3.WithPresignExpires(m.signedURLTTL))
	if err != nil {
		return "", fmt.Errorf("failed to presign PUT for key %q: %w", key, err)
	}

	return req.URL, nil
}

// DeleteObject deletes an object from S3.
func (m *Manager) DeleteObject(ctx context.Context, key string) error {
	_, err := m.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(m.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object %q: %w", key, err)
	}

	return nil
}

// DeleteObjects deletes multiple objects in a single request (max 1000).
func (m *Manager) DeleteObjects(ctx context.Context, keys []string) error {
	objects := make([]types.ObjectIdentifier, len(keys))
	for i, k := range keys {
		objects[i] = types.ObjectIdentifier{Key: aws.String(k)}
	}

	_, err := m.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(m.bucket),
		Delete: &types.Delete{Objects: objects},
	})
	if err != nil {
		return fmt.Errorf("failed to batch delete objects: %w", err)
	}

	return nil
}
