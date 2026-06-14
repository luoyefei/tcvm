package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <instance-id> <command>",
	Short: "Execute a command on a CVM instance via TAT",
	Long:  `Execute a shell command on the specified CVM instance using Tencent Automation Tools (TAT).`,
	Args:  cobra.MinimumNArgs(2),
	RunE:  runExec,
}

var (
	execType    string
	execTimeout time.Duration
	execWait    bool
)

func init() {
	rootCmd.AddCommand(execCmd)
	execCmd.Flags().StringVarP(&execType, "type", "t", "SHELL", "Command type: SHELL, POWERSHELL, BAT")
	execCmd.Flags().DurationVarP(&execTimeout, "timeout", "T", 60*time.Second, "Timeout for command execution")
	execCmd.Flags().BoolVarP(&execWait, "wait", "w", true, "Wait for command to complete")
}

func runExec(cmd *cobra.Command, args []string) error {
	instanceId := args[0]
	command := args[1]
	if len(args) > 2 {
		command = strings.Join(args[1:], " ")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	tatClient, err := client.TAT()
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Executing command on %s...\n", instanceId)

	invocationId, err := tatClient.RunCommand(context.Background(), []string{instanceId}, command, execType)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Invocation ID: %s\n", invocationId)

	if !execWait {
		fmt.Println(invocationId)
		return nil
	}

	fmt.Fprintln(os.Stderr, "Waiting for command to complete...")

	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()

	results, err := tatClient.WaitForCommand(ctx, invocationId, []string{instanceId}, execTimeout)
	if err != nil {
		return err
	}

	for _, r := range results {
		fmt.Fprintf(os.Stderr, "\n[%s] Status: %s | ExitCode: %d\n", r.InstanceId, r.Status, r.ExitCode)
		if r.Output != "" {
			fmt.Println(r.Output)
		}
	}

	return nil
}
