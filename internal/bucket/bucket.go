package bucket

import (
	"context"
	"fmt"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint  string
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
}

type Client struct {
	objects *minio.Client
	bucket  string
}

func New(config Config) (*Client, error) {
	if config.Endpoint == "" || config.Bucket == "" || config.Region == "" || config.AccessKey == "" || config.SecretKey == "" {
		return nil, fmt.Errorf("SAVES_ENDPOINT, SAVES_BUCKET, SAVES_REGION, AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY are required")
	}

	endpoint, secure := strings.TrimPrefix(config.Endpoint, "https://"), !strings.HasPrefix(config.Endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	objects, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure:       secure,
		Region:       config.Region,
		BucketLookup: minio.BucketLookupPath,
	})
	if err != nil {
		return nil, fmt.Errorf("create object storage client: %w", err)
	}

	return &Client{objects: objects, bucket: config.Bucket}, nil
}

func (c *Client) Purge(ctx context.Context, prefix string) error {
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for object := range c.objects.ListObjects(ctx, c.bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
		if object.Err != nil {
			return fmt.Errorf("list %s: %w", prefix, object.Err)
		}
		if err := c.objects.RemoveObject(ctx, c.bucket, object.Key, minio.RemoveObjectOptions{}); err != nil {
			return fmt.Errorf("remove %s: %w", object.Key, err)
		}
	}

	return nil
}
