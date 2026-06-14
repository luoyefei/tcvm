package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List CVM instances",
	RunE:  runList,
}

var (
	listKeyword string
	listIds     []string
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&listKeyword, "keyword", "k", "", "Filter by instance name keyword")
	listCmd.Flags().StringSliceVarP(&listIds, "id", "i", nil, "Filter by specific instance IDs")
}

func runList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	cvmClient, err := client.CVM()
	if err != nil {
		return err
	}

	instances, err := cvmClient.DescribeInstances(context.Background(), listIds, listKeyword)
	if err != nil {
		return err
	}

	if len(instances) == 0 {
		fmt.Println("No instances found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "INSTANCE ID\tNAME\tTYPE\tSTATUS\tPRIVATE IP\tPUBLIC IP\tZONE")
	for _, inst := range instances {
		publicIP := inst.PublicIP
		if publicIP == "" {
			publicIP = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			inst.InstanceId, inst.InstanceName, inst.InstanceType,
			strings.ToLower(inst.Status), inst.PrivateIP, publicIP, inst.Zone)
	}
	w.Flush()

	return nil
}
