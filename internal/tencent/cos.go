package tencent

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"tcvm/internal/config"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type COSClient struct {
	client *cos.Client
	bucket string
}

func (c *Client) COS() (*COSClient, error) {
	bucketName := c.cosBucket()
	region := c.cosRegion()
	if bucketName == "" || region == "" {
		return nil, fmt.Errorf("COS bucket not configured, set cos_bucket and cos_region in config")
	}

	u, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucketName, region))
	if err != nil {
		return nil, fmt.Errorf("invalid COS bucket format: %w", err)
	}

	cred := c.Credential()
	client := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cred.SecretId,
			SecretKey: cred.SecretKey,
		},
	})

	return &COSClient{client: client, bucket: bucketName}, nil
}

func (c *COSClient) UploadFile(ctx context.Context, key string, data []byte) error {
	_, err := c.client.Object.Put(ctx, key, bytes.NewReader(data), nil)
	if err != nil {
		return fmt.Errorf("failed to upload to COS: %w", err)
	}
	return nil
}

func (c *COSClient) GetPresignedURL(ctx context.Context, key string, expire time.Duration) (string, error) {
	cred := c.client.GetCredential()
	ak := cred.GetSecretId()
	sk := cred.GetSecretKey()

	presignedURL, err := c.client.Object.GetPresignedURL(ctx, http.MethodGet, key, ak, sk, expire, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return presignedURL.String(), nil
}

func (c *COSClient) DeleteObject(ctx context.Context, key string) error {
	_, err := c.client.Object.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete COS object: %w", err)
	}
	return nil
}

func (c *Client) cosBucket() string {
	return config.AppConfig.CosBucket
}

func (c *Client) cosRegion() string {
	return config.AppConfig.CosRegion
}
