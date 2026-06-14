package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <instance-id> <local-file> <remote-path>",
	Short: "Upload a file to a CVM instance",
	Long:  `Upload a local file to the specified path on a CVM instance using TAT or COS transfer.`,
	Args:  cobra.ExactArgs(3),
	RunE:  runUpload,
}

func init() {
	rootCmd.AddCommand(uploadCmd)
}

func runUpload(cmd *cobra.Command, args []string) error {
	instanceId := args[0]
	localPath := args[1]
	remotePath := args[2]

	client, err := newClient()
	if err != nil {
		return err
	}

	tatClient, err := client.TAT()
	if err != nil {
		return err
	}

	onStatus := func(status string) {
		fmt.Fprintf(os.Stderr, "[%s] ", status)
	}

	return uploadFile(context.Background(), instanceId, localPath, remotePath, tatClient, onStatus)
}
