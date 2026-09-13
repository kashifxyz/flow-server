package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Endpoint     string
	Region       string
	Bucket       string
	AccessKey    string
	SecretKey    string
	UsePathStyle bool
}

func (c Config) Enabled() bool {
	return c.Endpoint != "" && c.Bucket != "" && c.AccessKey != "" && c.SecretKey != ""
}

func New(ctx context.Context, cfg Config) (*s3.Client, error) {
	if !cfg.Enabled() {
		return nil, nil
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load s3 config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = cfg.UsePathStyle
	})
	return client, nil
}

func Ping(ctx context.Context, client *s3.Client, bucket string) error {
	if client == nil {
		return nil
	}
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		return fmt.Errorf("head bucket %s: %w", bucket, err)
	}
	return nil
}

// PresignPutURL returns a time-limited URL the client can PUT the object body to
// directly, so uploads never pass through the Go server (DEPENDENCIES.md §68).
func PresignPutURL(ctx context.Context, client *s3.Client, bucket, key, contentType string, ttl time.Duration) (string, error) {
	if client == nil {
		return "", fmt.Errorf("object storage is not configured")
	}
	presigner := s3.NewPresignClient(client)
	req, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("presign put %s: %w", key, err)
	}
	return req.URL, nil
}

// PresignGetURL returns a time-limited URL the client can GET the object body from directly.
func PresignGetURL(ctx context.Context, client *s3.Client, bucket, key string, ttl time.Duration) (string, error) {
	if client == nil {
		return "", fmt.Errorf("object storage is not configured")
	}
	presigner := s3.NewPresignClient(client)
	req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("presign get %s: %w", key, err)
	}
	return req.URL, nil
}
