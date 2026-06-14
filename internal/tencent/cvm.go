package tencent

import (
	"context"
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type CVMClient struct {
	client *cvm.Client
}

type Instance struct {
	InstanceId   string
	InstanceName string
	InstanceType string
	PrivateIP    string
	PublicIP     string
	Status       string
	Zone         string
	CreatedTime  string
}

func (c *Client) CVM() (*CVMClient, error) {
	prof := profile.NewClientProfile()
	prof.HttpProfile.Endpoint = "cvm.tencentcloudapi.com"

	client, err := cvm.NewClient(c.Credential(), c.Region(), prof)
	if err != nil {
		return nil, fmt.Errorf("failed to create cvm client: %w", err)
	}

	return &CVMClient{client: client}, nil
}

func (c *CVMClient) DescribeInstances(ctx context.Context, instanceIds []string, keyword string) ([]Instance, error) {
	request := cvm.NewDescribeInstancesRequest()
	request.Limit = int64Ptr(100)

	if len(instanceIds) > 0 {
		request.InstanceIds = make([]*string, len(instanceIds))
		for i, id := range instanceIds {
			request.InstanceIds[i] = stringPtr(id)
		}
	}

	if keyword != "" {
		request.Filters = []*cvm.Filter{
			{
				Name:   stringPtr("instance-name"),
				Values: []*string{stringPtr(keyword)},
			},
		}
	}

	response, err := c.client.DescribeInstancesWithContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	var instances []Instance
	for _, info := range response.Response.InstanceSet {
		inst := Instance{
			InstanceId:   deref(info.InstanceId),
			InstanceName: deref(info.InstanceName),
			InstanceType: deref(info.InstanceType),
			Status:       deref(info.InstanceState),
			Zone:         deref(info.Placement.Zone),
			CreatedTime:  deref(info.CreatedTime),
		}
		if len(info.PrivateIpAddresses) > 0 {
			inst.PrivateIP = deref(info.PrivateIpAddresses[0])
		}
		if len(info.PublicIpAddresses) > 0 {
			inst.PublicIP = deref(info.PublicIpAddresses[0])
		}
		instances = append(instances, inst)
	}

	return instances, nil
}

func (c *CVMClient) GetVncUrl(ctx context.Context, instanceId string) (string, error) {
	request := cvm.NewDescribeInstanceVncUrlRequest()
	request.InstanceId = stringPtr(instanceId)

	response, err := c.client.DescribeInstanceVncUrlWithContext(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to get vnc url: %w", err)
	}

	vncUrl := fmt.Sprintf("https://img.qcloud.com/qcloud/app/active_vnc/index.html?InstanceVncUrl=%s",
		deref(response.Response.InstanceVncUrl))

	return vncUrl, nil
}
