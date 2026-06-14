package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"tcvm/internal/tencent"
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

type connectSession struct {
	instanceId string
	cwd        string
	timeout    time.Duration
	reader     *bufio.Reader
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

	session := &connectSession{
		instanceId: instanceId,
		cwd:        "/root",
		timeout:    connectTimeout,
	}

	fmt.Fprintf(os.Stderr, "Connecting to %s...\n\n", instanceId)
	return session.runInteractive(tatClient, bufio.NewReader(os.Stdin))
}

func (s *connectSession) runInteractive(tatClient *tencent.TATClient, reader *bufio.Reader) error {
	fmt.Fprintf(os.Stderr, "Connected to %s\n", s.instanceId)
	fmt.Fprintln(os.Stderr, "Type 'exit' or press Ctrl+D to quit.")
	fmt.Fprintln(os.Stderr, "Special: upfile <local-path> <remote-path>")
	fmt.Fprintln(os.Stderr, "")

	s.reader = reader
	for {
		prompt := fmt.Sprintf("%s:%s$ ", s.instanceId, s.cwd)
		fmt.Print(prompt)

		line, err := s.reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "\nDisconnected.")
			return nil
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if line == "exit" || line == "quit" {
			fmt.Fprintln(os.Stderr, "Disconnected.")
			return nil
		}

		if strings.HasPrefix(line, "upfile") {
			if err := s.handleUpload(tatClient, line); err != nil {
				fmt.Fprintf(os.Stderr, "upload error: %v\n", err)
			}
			continue
		}

		if err := s.execute(tatClient, line); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}
}

func (s *connectSession) handleUpload(tatClient *tencent.TATClient, line string) error {
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return fmt.Errorf("usage: upfile <local-path> <remote-path>")
	}

	localPath := parts[1]
	remotePath := parts[2]

	onStatus := func(status string) {
		fmt.Fprintf(os.Stderr, "[%s] ", status)
	}

	return uploadFile(context.Background(), s.instanceId, localPath, remotePath, tatClient, onStatus)
}

func (s *connectSession) execute(tatClient *tencent.TATClient, line string) error {
	isCd := strings.HasPrefix(line, "cd ")
	command := s.buildCommand(line)

	invocationId, err := tatClient.RunCommand(context.Background(), []string{s.instanceId}, command, "SHELL")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	results, err := tatClient.WaitForCommand(ctx, invocationId, []string{s.instanceId}, s.timeout)
	if err != nil {
		return err
	}

	if len(results) > 0 {
		r := results[0]
		if r.Status == "FAILED" && r.ExitCode != 0 {
			fmt.Fprintf(os.Stderr, "exit code: %d\n", r.ExitCode)
		}
		if r.Output != "" {
			output := strings.TrimSpace(r.Output)
			if isCd && r.Status == "SUCCESS" {
				s.cwd = output
			} else {
				fmt.Println(output)
			}
		}
	}

	return nil
}

func (s *connectSession) buildCommand(line string) string {
	if strings.HasPrefix(line, "cd ") {
		target := strings.TrimSpace(strings.TrimPrefix(line, "cd "))
		if target == "~" || target == "" {
			target = "/root"
		}
		return fmt.Sprintf("cd %s && pwd", target)
	}

	return fmt.Sprintf("cd %s && %s", s.cwd, line)
}
