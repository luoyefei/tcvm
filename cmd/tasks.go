package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks <invocation-id>",
	Short: "Query command execution results",
	Long:  `Query the status and output of a TAT command execution by invocation ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTasks,
}

var tasksInstanceId string

func init() {
	rootCmd.AddCommand(tasksCmd)
	tasksCmd.Flags().StringVarP(&tasksInstanceId, "instance", "i", "", "Filter by specific instance ID")
}

func runTasks(cmd *cobra.Command, args []string) error {
	invocationId := args[0]

	client, err := newClient()
	if err != nil {
		return err
	}

	tatClient, err := client.TAT()
	if err != nil {
		return err
	}

	var instanceIds []string
	if tasksInstanceId != "" {
		instanceIds = []string{tasksInstanceId}
	}

	results, err := tatClient.DescribeInvocationTasks(context.Background(), invocationId, instanceIds)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	for _, r := range results {
		fmt.Printf("[%s] Status: %s | ExitCode: %d\n", r.InstanceId, r.Status, r.ExitCode)
		if r.Output != "" {
			fmt.Println(r.Output)
		}
		fmt.Println()
	}

	return nil
}
