package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var vncCmd = &cobra.Command{
	Use:   "vnc <instance-id>",
	Short: "Get the VNC console URL for a CVM instance",
	Long: `Retrieve a VNC URL for graphical console access to the specified CVM instance.

The URL is valid for a limited time and provides browser-based console access,
useful when the instance has no network connectivity at all.`,
	Args: cobra.ExactArgs(1),
	RunE: runVNC,
}

func init() {
	rootCmd.AddCommand(vncCmd)
}

func runVNC(cmd *cobra.Command, args []string) error {
	instanceId := args[0]

	client, err := newClient()
	if err != nil {
		return err
	}

	cvmClient, err := client.CVM()
	if err != nil {
		return err
	}

	url, err := cvmClient.GetVncUrl(context.Background(), instanceId)
	if err != nil {
		return err
	}

	fmt.Println(url)
	return nil
}
