package tencent

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tat "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tat/v20201028"
)

type TATClient struct {
	client *tat.Client
}

type CommandResult struct {
	InvocationId string
	InstanceId   string
	TaskId       string
	Status       string
	Output       string
	ExitCode     int64
}

func (c *Client) TAT() (*TATClient, error) {
	prof := profile.NewClientProfile()
	prof.HttpProfile.Endpoint = "tat.tencentcloudapi.com"

	client, err := tat.NewClient(c.Credential(), c.Region(), prof)
	if err != nil {
		return nil, fmt.Errorf("failed to create tat client: %w", err)
	}

	return &TATClient{client: client}, nil
}

func (t *TATClient) RunCommand(ctx context.Context, instanceIds []string, command string, commandType string) (string, error) {
	request := tat.NewRunCommandRequest()
	request.Content = stringPtr(base64.StdEncoding.EncodeToString([]byte(command)))
	request.InstanceIds = make([]*string, len(instanceIds))
	for i, id := range instanceIds {
		request.InstanceIds[i] = stringPtr(id)
	}

	if commandType == "" {
		commandType = "SHELL"
	}
	request.CommandType = stringPtr(commandType)

	response, err := t.client.RunCommandWithContext(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	return deref(response.Response.InvocationId), nil
}

func (t *TATClient) DescribeInvocationTasks(ctx context.Context, invocationId string, instanceIds []string) ([]CommandResult, error) {
	request := tat.NewDescribeInvocationTasksRequest()
	request.Filters = []*tat.Filter{
		{
			Name:   stringPtr("invocation-id"),
			Values: []*string{stringPtr(invocationId)},
		},
	}

	if len(instanceIds) > 0 {
		request.Filters = append(request.Filters, &tat.Filter{
			Name:   stringPtr("instance-id"),
			Values: strSlicePtr(instanceIds),
		})
	}

	response, err := t.client.DescribeInvocationTasksWithContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to describe invocation tasks: %w", err)
	}

	var results []CommandResult
	for _, task := range response.Response.InvocationTaskSet {
		result := CommandResult{
			InvocationId: invocationId,
			InstanceId:   deref(task.InstanceId),
			TaskId:       deref(task.InvocationTaskId),
			Status:       deref(task.TaskStatus),
		}
		if task.TaskResult != nil {
			result.ExitCode = derefInt64(task.TaskResult.ExitCode)
			if task.TaskResult.Output != nil {
				out, _ := base64.StdEncoding.DecodeString(deref(task.TaskResult.Output))
				result.Output = string(out)
			}
		}
		results = append(results, result)
	}

	return results, nil
}

func (t *TATClient) WaitForCommand(ctx context.Context, invocationId string, instanceIds []string, timeout time.Duration) ([]CommandResult, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		results, err := t.DescribeInvocationTasks(ctx, invocationId, instanceIds)
		if err != nil {
			return nil, err
		}

		allFinished := true
		for _, r := range results {
			if r.Status != "SUCCESS" && r.Status != "FAILED" && r.Status != "TIMEOUT" && r.Status != "CANCELLED" {
				allFinished = false
				break
			}
		}

		if allFinished {
			return results, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	return nil, fmt.Errorf("command execution timeout after %v", timeout)
}

func (t *TATClient) UploadFile(ctx context.Context, instanceId string, fileContent []byte, remotePath string, overwrite bool, onStatus func(string)) error {
	const maxUploadSize = 24 * 1024
	if len(fileContent) > maxUploadSize {
		return fmt.Errorf("file too large (%d bytes). TAT direct upload limit is ~%d bytes. Use COS or other transfer methods for large files", len(fileContent), maxUploadSize)
	}

	request := tat.NewRunCommandRequest()

	encodedContent := base64.StdEncoding.EncodeToString(fileContent)
	script := fmt.Sprintf(`cat << 'EOF' | base64 -d > %s
%s
EOF`, remotePath, encodedContent)
	if overwrite {
		script = fmt.Sprintf(`cat << 'EOF' | base64 -d > %s
%s
EOF`, remotePath, encodedContent)
	} else {
		script = fmt.Sprintf(`[ -f %s ] || cat << 'EOF' | base64 -d > %s
%s
EOF`, remotePath, remotePath, encodedContent)
	}

	request.Content = stringPtr(base64.StdEncoding.EncodeToString([]byte(script)))
	request.InstanceIds = []*string{stringPtr(instanceId)}
	request.CommandType = stringPtr("SHELL")

	response, err := t.client.RunCommandWithContext(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	invocationId := deref(response.Response.InvocationId)

	if onStatus != nil {
		onStatus("PENDING")
	}

	results, err := t.waitForCommandWithProgress(ctx, invocationId, []string{instanceId}, 60*time.Second, onStatus)
	if err != nil {
		return err
	}

	if len(results) > 0 && results[0].Status != "SUCCESS" {
		return fmt.Errorf("upload failed: status=%s, output=%s", results[0].Status, results[0].Output)
	}

	return nil
}

func (t *TATClient) waitForCommandWithProgress(ctx context.Context, invocationId string, instanceIds []string, timeout time.Duration, onStatus func(string)) ([]CommandResult, error) {
	deadline := time.Now().Add(timeout)
	lastStatus := ""

	for time.Now().Before(deadline) {
		results, err := t.DescribeInvocationTasks(ctx, invocationId, instanceIds)
		if err != nil {
			return nil, err
		}

		allFinished := true
		for _, r := range results {
			if r.Status != lastStatus && onStatus != nil {
				onStatus(r.Status)
				lastStatus = r.Status
			}
			if r.Status != "SUCCESS" && r.Status != "FAILED" && r.Status != "TIMEOUT" && r.Status != "CANCELLED" {
				allFinished = false
			}
		}

		if allFinished {
			return results, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	return nil, fmt.Errorf("command execution timeout after %v", timeout)
}
