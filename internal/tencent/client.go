package tencent

import (
	"fmt"
	"tcvm/internal/config"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

type Client struct {
	regionOverride string
}

func NewClient() (*Client, error) {
	if !config.IsConfigured() {
		return nil, fmt.Errorf("tencent cloud credentials not configured, run 'tcvm config' first")
	}

	return &Client{}, nil
}

func NewClientWithRegion(region string) (*Client, error) {
	c, err := NewClient()
	if err != nil {
		return nil, err
	}
	c.regionOverride = region
	return c, nil
}

func (c *Client) Credential() *common.Credential {
	return common.NewCredential(
		config.AppConfig.SecretId,
		config.AppConfig.SecretKey,
	)
}

func (c *Client) Region() string {
	if c.regionOverride != "" {
		return c.regionOverride
	}
	return config.AppConfig.Region
}
