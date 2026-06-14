package cmd

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tcvm/internal/tencent"
	"time"
)

const maxTATUploadSize = 24 * 1024

func uploadFile(ctx context.Context, instanceId string, localPath, remotePath string, tatClient *tencent.TATClient, onStatus func(string)) error {
	fileContent, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %w", err)
	}

	if len(fileContent) <= maxTATUploadSize {
		fmt.Fprintf(os.Stderr, "Uploading %s -> %s (%d bytes) ", localPath, remotePath, len(fileContent))
		if err := tatClient.UploadFile(ctx, instanceId, fileContent, remotePath, true, onStatus); err != nil {
			fmt.Fprintln(os.Stderr)
			return err
		}
		fmt.Fprintln(os.Stderr, "done.")
		return nil
	}

	return uploadViaCOS(ctx, instanceId, localPath, remotePath, fileContent, tatClient)
}

func uploadViaCOS(ctx context.Context, instanceId string, localPath, remotePath string, fileContent []byte, tatClient *tencent.TATClient) error {
	client, err := tencent.NewClient()
	if err != nil {
		return err
	}

	cosClient, err := client.COS()
	if err != nil {
		return fmt.Errorf("COS not available: %w. File too large for TAT direct upload (%d bytes > %d bytes)", err, len(fileContent), maxTATUploadSize)
	}

	localMD5 := md5Hash(fileContent)
	baseName := filepath.Base(localPath)
	cosKey := fmt.Sprintf("tcvm-uploads/%d-%s", time.Now().Unix(), baseName)

	fmt.Fprintf(os.Stderr, "File too large for TAT (%d bytes). Using COS transfer...\n", len(fileContent))
	fmt.Fprintf(os.Stderr, "Uploading to COS [%s]... ", cosKey)

	if err := cosClient.UploadFile(ctx, cosKey, fileContent); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "done.")

	url, err := cosClient.GetPresignedURL(ctx, cosKey, 5*time.Minute)
	if err != nil {
		_ = cosClient.DeleteObject(ctx, cosKey)
		return err
	}

	fmt.Fprintf(os.Stderr, "Downloading to instance %s... ", instanceId)

	downloadCmd := fmt.Sprintf(`wget -q -O %s "%s" 2>/dev/null || curl -sL -o %s "%s"`, remotePath, url, remotePath, url)
	invocationId, err := tatClient.RunCommand(ctx, []string{instanceId}, downloadCmd, "SHELL")
	if err != nil {
		_ = cosClient.DeleteObject(ctx, cosKey)
		return err
	}

	results, err := tatClient.WaitForCommand(ctx, invocationId, []string{instanceId}, 120*time.Second)
	if err != nil {
		_ = cosClient.DeleteObject(ctx, cosKey)
		return err
	}

	if len(results) > 0 && results[0].Status != "SUCCESS" {
		_ = cosClient.DeleteObject(ctx, cosKey)
		return fmt.Errorf("download failed: status=%s, output=%s", results[0].Status, results[0].Output)
	}
	fmt.Fprintln(os.Stderr, "done.")

	fmt.Fprintf(os.Stderr, "Verifying MD5... ")
	md5Cmd := fmt.Sprintf("md5sum %s | awk '{print $1}'", remotePath)
	invocationId, err = tatClient.RunCommand(ctx, []string{instanceId}, md5Cmd, "SHELL")
	if err != nil {
		_ = cosClient.DeleteObject(ctx, cosKey)
		return err
	}

	results, err = tatClient.WaitForCommand(ctx, invocationId, []string{instanceId}, 60*time.Second)
	if err != nil {
		_ = cosClient.DeleteObject(ctx, cosKey)
		return err
	}

	remoteMD5 := ""
	if len(results) > 0 && results[0].Output != "" {
		remoteMD5 = strings.TrimSpace(results[0].Output)
	}

	if remoteMD5 != localMD5 {
		_ = cosClient.DeleteObject(ctx, cosKey)
		return fmt.Errorf("MD5 mismatch: local=%s remote=%s", localMD5, remoteMD5)
	}
	fmt.Fprintln(os.Stderr, "OK.")

	fmt.Fprintf(os.Stderr, "Cleaning up COS temp file... ")
	if err := cosClient.DeleteObject(ctx, cosKey); err != nil {
		fmt.Fprintln(os.Stderr, "warning: failed to delete COS object:", err)
	} else {
		fmt.Fprintln(os.Stderr, "done.")
	}

	return nil
}

func md5Hash(data []byte) string {
	h := md5.Sum(data)
	return hex.EncodeToString(h[:])
}
