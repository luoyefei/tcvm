package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"tcvm/internal/tencent"
	"time"
)

// pseudoSession drives an interactive pseudo-shell against a CVM instance via
// TAT. It backs both the `connect` and `shell` commands; the only behavioural
// difference is whether a user is shown in the prompt and used for `~`.
type pseudoSession struct {
	instanceId string
	cwd        string
	user       string // empty -> no "user@" in prompt
	timeout    time.Duration
	reader     *bufio.Reader
}

// shellQuote wraps s in single quotes so it can be embedded safely in a shell
// command, protecting against spaces and metacharacters in paths.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func (s *pseudoSession) prompt() string {
	if s.user != "" {
		return fmt.Sprintf("%s@%s:%s$ ", s.user, s.instanceId, s.cwd)
	}
	return fmt.Sprintf("%s:%s$ ", s.instanceId, s.cwd)
}

func (s *pseudoSession) homeDir() string {
	if s.user == "" || s.user == "root" {
		return "/root"
	}
	return "/home/" + s.user
}

func (s *pseudoSession) runInteractive(tatClient *tencent.TATClient) error {
	fmt.Fprintf(os.Stderr, "Connected to %s\n", s.instanceId)
	fmt.Fprintln(os.Stderr, "Type 'exit' or press Ctrl+D to quit.")
	fmt.Fprintln(os.Stderr, "Special: upfile <local-path> <remote-path>")
	fmt.Fprintln(os.Stderr, "Note: ~1-3s latency per command. vim/top/sudo password are NOT supported.")
	fmt.Fprintln(os.Stderr, "")

	for {
		fmt.Print(s.prompt())

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

func (s *pseudoSession) handleUpload(tatClient *tencent.TATClient, line string) error {
	parts := strings.Fields(line)
	if len(parts) != 3 {
		return fmt.Errorf("usage: upfile <local-path> <remote-path>")
	}

	onStatus := func(status string) {
		fmt.Fprintf(os.Stderr, "[%s] ", status)
	}

	return uploadFile(context.Background(), s.instanceId, parts[1], parts[2], tatClient, onStatus)
}

func (s *pseudoSession) execute(tatClient *tencent.TATClient, line string) error {
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

func (s *pseudoSession) buildCommand(line string) string {
	if strings.HasPrefix(line, "cd ") {
		// Leave the target unquoted so ~, $VARs and globs still expand.
		target := strings.TrimSpace(strings.TrimPrefix(line, "cd "))
		if target == "~" || target == "" {
			target = s.homeDir()
		}
		return fmt.Sprintf("cd %s && pwd", target)
	}

	// cwd comes from a real `pwd` result, so quote it against spaces.
	return fmt.Sprintf("cd %s && %s", shellQuote(s.cwd), line)
}
