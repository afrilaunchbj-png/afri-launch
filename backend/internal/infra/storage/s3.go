package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3 stocke les objets dans un stockage objet compatible S3 (Neon Object
// Storage en production). Remplace LocalStorage (disque éphémère sur Railway).
type S3 struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
}

// NewS3 construit un stockage S3-compatible. endpoint vide = AWS S3 ;
// pathStyle force l'adressage par chemin (requis par la plupart des
// endpoints compatibles S3).
func NewS3(ctx context.Context, endpoint, region, accessKeyID, secretAccessKey, bucket string, pathStyle bool) (*S3, error) {
	if bucket == "" {
		return nil, errors.New("storage: bucket S3 manquant")
	}
	if region == "" {
		region = "us-east-1"
	}

	opts := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(region)}
	if accessKeyID != "" && secretAccessKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("storage: config aws: %w", err)
	}

	clientOpts := []func(*s3.Options){}
	if endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(strings.TrimRight(endpoint, "/"))
		})
	}
	if pathStyle {
		clientOpts = append(clientOpts, func(o *s3.Options) { o.UsePathStyle = true })
	}

	return &S3{client: s3.NewFromConfig(awsCfg, clientOpts...), presign: s3.NewPresignClient(s3.NewFromConfig(awsCfg, clientOpts...)), bucket: bucket}, nil
}

// Put écrit data sous la clé donnée.
func (s *S3) Put(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(strings.TrimPrefix(key, "/")),
		Body:        strings.NewReader(string(data)),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("storage s3 put %q: %w", key, err)
	}
	return nil
}

// Get lit l'objet sous la clé donnée.
func (s *S3) Get(ctx context.Context, key string) ([]byte, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(strings.TrimPrefix(key, "/")),
	})
	if err != nil {
		return nil, fmt.Errorf("storage s3 get %q: %w", key, err)
	}
	defer out.Body.Close()
	data, err := io.ReadAll(out.Body)
	if err != nil {
		return nil, fmt.Errorf("storage s3 read %q: %w", key, err)
	}
	return data, nil
}

// SignedURL produit une URL présignée temporairement publique (GET) —
// utilisée pour fournir les creatives aux plateformes publicitaires.
func (s *S3) SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(strings.TrimPrefix(key, "/")),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("storage s3 presign %q: %w", key, err)
	}
	return req.URL, nil
}
