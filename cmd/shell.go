package cmd

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell <instance-id>",
	Short: "Start an interactive shell session on a CVM instance",
	Long: `Start an interactive pseudo-shell on the specified CVM instance via TAT.

This provides a REPL-like experience where commands are sent through TAT API.
Note: This is not a real SSH session. Each command is an independent API call,
so there is noticeable latency (~1-3s per command). Interactive programs like vim/top
are not supported.`,
	Args: cobra.ExactArgs(1),
	RunE: runShell,
}

var (
	shellTimeout time.Duration
	shellUser    string
)

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.Flags().DurationVarP(&shellTimeout, "timeout", "T", 60*time.Second, "Default timeout per command")
	shellCmd.Flags().StringVarP(&shellUser, "user", "u", "root", "Default user to run commands as")
}

func runShell(cmd *cobra.Command, args []string) error {
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
		user:       shellUser,
		timeout:    shellTimeout,
		reader:     bufio.NewReader(os.Stdin),
	}

	fmt.Fprintf(os.Stderr, "Connected to %s via TAT (pseudo-shell)\n", instanceId)
	return session.runInteractive(tatClient)
}
