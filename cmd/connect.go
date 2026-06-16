package cmd

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect <instance-id>",
	Short: "Connect to a CVM instance and start an interactive shell",
	Long: `Connect to the specified CVM instance via TAT and start an interactive pseudo-shell.

Each command you type is sent to the instance through Tencent Automation Tools (TAT).
The CLI maintains your current directory state locally so cd/ls/cat etc feel natural.
Special command:
  upfile <local-path> <remote-path>  Upload a file
Note: ~1-3s latency per command. vim/top/sudo password are NOT supported.`,
	Args: cobra.ExactArgs(1),
	RunE: runConnect,
}

var connectTimeout time.Duration

func init() {
	rootCmd.AddCommand(connectCmd)
	connectCmd.Flags().DurationVarP(&connectTimeout, "timeout", "T", 60*time.Second, "Default timeout per command")
}

func runConnect(cmd *cobra.Command, args []string) error {
	instanceId := args[0]

	client, err := newClient()
	if err != nil {
		return err
	}

	tatClient, err := client.TAT()
	if err != nil {
		return err
	}

	session := &pseudoSession{
		instanceId: instanceId,
		cwd:        "/root",
		timeout:    connectTimeout,
		reader:     bufio.NewReader(os.Stdin),
	}

	fmt.Fprintf(os.Stderr, "Connecting to %s...\n\n", instanceId)
	return session.runInteractive(tatClient)
}
