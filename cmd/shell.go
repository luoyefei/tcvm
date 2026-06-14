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

type shellSession struct {
	instanceId string
	cwd        string
	user       string
	timeout    time.Duration
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

	session := &shellSession{
		instanceId: instanceId,
		cwd:        "/root",
		user:       shellUser,
		timeout:    shellTimeout,
	}

	fmt.Fprintf(os.Stderr, "Connected to %s via TAT (pseudo-shell)\n", instanceId)
	fmt.Fprintln(os.Stderr, "Type 'exit' or press Ctrl+D to quit.")
	fmt.Fprintln(os.Stderr, "Note: ~1-3s latency per command. vim/top/sudo password are NOT supported.")
	fmt.Fprintln(os.Stderr, "")

	reader := bufio.NewReader(os.Stdin)
	for {
		prompt := fmt.Sprintf("%s@%s:%s$ ", session.user, session.instanceId, session.cwd)
		fmt.Print(prompt)

		line, err := reader.ReadString('\n')
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

		if err := session.execute(tatClient, line); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}
}

func (s *shellSession) execute(tatClient *tencent.TATClient, line string) error {
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

func (s *shellSession) buildCommand(line string) string {
	if strings.HasPrefix(line, "cd ") {
		target := strings.TrimSpace(strings.TrimPrefix(line, "cd "))
		if target == "~" || target == "" {
			target = fmt.Sprintf("/home/%s", s.user)
			if s.user == "root" {
				target = "/root"
			}
		}
		return fmt.Sprintf("cd %s && pwd", target)
	}

	return fmt.Sprintf("cd %s && %s", s.cwd, line)
}
