package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"tcvm/internal/config"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	accentColor   = lipgloss.Color("#00D4AA")
	primaryColor  = lipgloss.Color("#0A2540")
	textColor     = lipgloss.Color("#E3E8EE")
	dimColor      = lipgloss.Color("#6B7280")
	successColor  = lipgloss.Color("#10B981")
	warningColor  = lipgloss.Color("#F59E0B")
	errorColor    = lipgloss.Color("#EF4444")
)

var rootCmd = &cobra.Command{
	Use:   "tcvm",
	Short: "A CLI tool for managing Tencent Cloud CVM instances",
	Long: `tcvm is a command-line tool that helps you manage Tencent Cloud CVM instances
through APIs, especially useful for instances without public network access.`,
	RunE: runWorkbench,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "config" {
			return nil
		}
		if configFile != "" {
			config.SetConfigFile(configFile)
		}
		if err := config.InitConfig(); err != nil {
			return err
		}
		if !config.IsConfigured() {
			return fmt.Errorf("tencent cloud credentials not configured. Run 'tcvm config' to set up")
		}
		return nil
	},
	SilenceUsage: true,
}

func Execute() {
	rootCmd.SetHelpFunc(renderHelp)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func renderHelp(cmd *cobra.Command, args []string) {
	bannerColor := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(textColor)
	dimStyle := lipgloss.NewStyle().Foreground(dimColor)
	keyStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	sectionStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	cmdStyle := lipgloss.NewStyle().Foreground(textColor)
	cmdDescStyle := lipgloss.NewStyle().Foreground(dimColor)

	fmt.Println()
	fmt.Println(bannerColor.Render("  ████████╗ ██████╗██╗   ██╗███╗   ███╗"))
	fmt.Println(bannerColor.Render("  ╚══██╔══╝██╔════╝██║   ██║████╗ ████║"))
	fmt.Println(bannerColor.Render("     ██║   ██║     ██║   ██║██╔████╔██║"))
	fmt.Println(bannerColor.Render("     ██║   ██║     ██║   ██║██║╚██╔╝██║"))
	fmt.Println(bannerColor.Render("     ██║   ╚██████╗╚██████╔╝██║ ╚═╝ ██║"))
	fmt.Println(bannerColor.Render("     ╚═╝    ╚═════╝ ╚═════╝ ╚═╝     ╚═╝"))
	fmt.Println(dimStyle.Render("      Tencent Cloud VM Manager"))
	fmt.Println("  " + dimStyle.Render(strings.Repeat("━", 60)))
	fmt.Println()

	fmt.Println(descStyle.Render("  A CLI tool for managing Tencent Cloud CVM instances through APIs,"))
	fmt.Println(descStyle.Render("  especially useful for instances without public network access."))
	fmt.Println()

	fmt.Println(sectionStyle.Render("  USAGE"))
	fmt.Println(dimStyle.Render("  tcvm [command] [flags]"))
	fmt.Println(dimStyle.Render("  tcvm                              Launch interactive workbench"))
	fmt.Println()

	fmt.Println(sectionStyle.Render("  COMMANDS"))
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	for _, c := range cmd.Commands() {
		if c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		fmt.Fprintf(w, "  %s\t%s\n",
			keyStyle.Render(c.Name()),
			cmdDescStyle.Render(c.Short))
	}
	w.Flush()
	fmt.Println()

	fmt.Println(sectionStyle.Render("  FLAGS"))
	w2 := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	seen := make(map[string]bool)
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if seen[f.Name] {
			return
		}
		seen[f.Name] = true
		fmt.Fprintf(w2, "  %s\t%s\n",
			keyStyle.Render("--"+f.Name),
			cmdDescStyle.Render(f.Usage))
	})
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if seen[f.Name] {
			return
		}
		seen[f.Name] = true
		fmt.Fprintf(w2, "  %s\t%s\n",
			keyStyle.Render("--"+f.Name),
			cmdDescStyle.Render(f.Usage))
	})
	w2.Flush()
	fmt.Println()

	fmt.Println(sectionStyle.Render("  EXAMPLES"))
	examples := []string{
		"  tcvm                              # Launch interactive menu",
		"  tcvm config                       # Configure credentials",
		"  tcvm list                         # List instances",
		"  tcvm connect ins-xxx              # Interactive shell",
		"  tcvm exec ins-xxx 'df -h'         # Execute command",
		"  tcvm upload ins-xxx ./f.txt /tmp  # Upload file",
	}
	for _, ex := range examples {
		parts := strings.SplitN(ex, "#", 2)
		if len(parts) == 2 {
			fmt.Println(cmdStyle.Render(parts[0]) + dimStyle.Render("#"+parts[1]))
		} else {
			fmt.Println(cmdStyle.Render(ex))
		}
	}
	fmt.Println()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/.tcvm/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&regionFlag, "region", "r", "", "Override region (e.g., ap-guangzhou)")
}

var (
	configFile string
	regionFlag string
)

func renderBanner() {
	bannerColor := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	subtitleColor := lipgloss.NewStyle().Foreground(dimColor)
	dividerColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#1E3A5F"))

	fmt.Println()
	fmt.Println(bannerColor.Render("  ████████╗ ██████╗██╗   ██╗███╗   ███╗"))
	fmt.Println(bannerColor.Render("  ╚══██╔══╝██╔════╝██║   ██║████╗ ████║"))
	fmt.Println(bannerColor.Render("     ██║   ██║     ██║   ██║██╔████╔██║"))
	fmt.Println(bannerColor.Render("     ██║   ██║     ██║   ██║██║╚██╔╝██║"))
	fmt.Println(bannerColor.Render("     ██║   ╚██████╗╚██████╔╝██║ ╚═╝ ██║"))
	fmt.Println(bannerColor.Render("     ╚═╝    ╚═════╝ ╚═════╝ ╚═╝     ╚═╝"))
	fmt.Println(subtitleColor.Render("      Tencent Cloud VM Manager  v1.0"))
	fmt.Println("  " + dividerColor.Render(strings.Repeat("━", 60)))
}

func renderMenu() {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#1E3A5F")).
		Padding(1, 2).
		Width(58)

	titleStyle := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		MarginBottom(1)

	itemStyle := lipgloss.NewStyle().
		Foreground(textColor).
		PaddingLeft(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true)

	items := []struct {
		key  string
		desc string
	}{
		{"1", "List instances"},
		{"2", "Connect to instance"},
		{"3", "Execute command"},
		{"4", "Upload file"},
		{"5", "View tasks"},
		{"6", "Configure"},
		{"h", "Help"},
		{"q", "Quit"},
	}

	var content strings.Builder
	content.WriteString(titleStyle.Render("  MAIN MENU"))
	content.WriteString("\n\n")

	for _, item := range items {
		content.WriteString(fmt.Sprintf("  %s  %s\n",
			keyStyle.Render("["+item.key+"]"),
			itemStyle.Render(item.desc)))
	}

	fmt.Println(boxStyle.Render(content.String()))
}

func renderStatusBar() {
	barStyle := lipgloss.NewStyle().
		Background(primaryColor).
		Foreground(textColor).
		Padding(0, 2).
		Width(60)

	region := config.AppConfig.Region
	if regionFlag != "" {
		region = regionFlag
	}

	status := fmt.Sprintf("  Region: %s  |  Auth: %s***%s",
		lipgloss.NewStyle().Foreground(accentColor).Render(region),
		config.AppConfig.SecretId[:8],
		config.AppConfig.SecretId[len(config.AppConfig.SecretId)-4:],
	)

	fmt.Println(barStyle.Render(status))
	fmt.Println()
}

func runWorkbench(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return cmd.Help()
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		renderBanner()
		renderMenu()
		renderStatusBar()

		promptStyle := lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

		fmt.Print(promptStyle.Render("  Select ") + lipgloss.NewStyle().Foreground(dimColor).Render("▸ "))

		input, err := reader.ReadString('\n')
		if err != nil {
			return nil
		}

		choice := strings.TrimSpace(strings.ToLower(input))
		if choice == "" {
			continue
		}
		if choice == "q" || choice == "quit" || choice == "exit" {
			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Foreground(dimColor).Render("  Goodbye."))
			fmt.Println()
			return nil
		}

		switch choice {
		case "1", "list":
			if err := workbenchList(); err != nil {
				renderError(err)
			}
		case "2", "connect":
			if err := workbenchConnect(reader); err != nil {
				renderError(err)
			}
		case "3", "exec":
			if err := workbenchExec(reader); err != nil {
				renderError(err)
			}
		case "4", "upload":
			if err := workbenchUpload(reader); err != nil {
				renderError(err)
			}
		case "5", "tasks":
			if err := workbenchTasks(reader); err != nil {
				renderError(err)
			}
		case "6", "config":
			if err := runConfig(nil, nil); err != nil {
				renderError(err)
			}
		case "h", "help":
			cmd.Help()
		default:
			renderError(fmt.Errorf("invalid choice: %s", choice))
		}
	}
}

func renderError(err error) {
	fmt.Println()
	errBox := lipgloss.NewStyle().
		Foreground(errorColor).
		PaddingLeft(2).
		Render("✗ " + err.Error())
	fmt.Println(errBox)
	fmt.Println()
}

func renderSuccess(msg string) {
	fmt.Println()
	successBox := lipgloss.NewStyle().
		Foreground(successColor).
		PaddingLeft(2).
		Render("✓ " + msg)
	fmt.Println(successBox)
	fmt.Println()
}

func workbenchList() error {
	client, err := newClient()
	if err != nil {
		return err
	}

	cvmClient, err := client.CVM()
	if err != nil {
		return err
	}

	instances, err := cvmClient.DescribeInstances(context.Background(), nil, "")
	if err != nil {
		return err
	}

	if len(instances) == 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Foreground(dimColor).PaddingLeft(2).Render("No instances found."))
		fmt.Println()
		return nil
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		MarginBottom(1).
		PaddingLeft(2)

	fmt.Println()
	fmt.Println(titleStyle.Render("INSTANCES"))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	header := lipgloss.NewStyle().Foreground(dimColor).Render
	fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n",
		header("#"), header("INSTANCE ID"), header("NAME"),
		header("STATUS"), header("PRIVATE IP"))

	for i, inst := range instances {
		statusColor := successColor
		if inst.Status != "RUNNING" {
			statusColor = warningColor
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n",
			lipgloss.NewStyle().Foreground(accentColor).Render(fmt.Sprintf("%d", i+1)),
			inst.InstanceId,
			inst.InstanceName,
			lipgloss.NewStyle().Foreground(statusColor).Render(strings.ToLower(inst.Status)),
			inst.PrivateIP)
	}
	w.Flush()
	fmt.Println()
	return nil
}

func workbenchConnect(reader *bufio.Reader) error {
	if err := workbenchList(); err != nil {
		return err
	}

	promptStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	fmt.Print(promptStyle.Render("  Select instance to connect (number): "))

	input, err := reader.ReadString('\n')
	if err != nil {
		return nil
	}

	idx, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return fmt.Errorf("invalid number")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	cvmClient, err := client.CVM()
	if err != nil {
		return err
	}

	instances, err := cvmClient.DescribeInstances(context.Background(), nil, "")
	if err != nil {
		return err
	}

	if idx < 1 || idx > len(instances) {
		return fmt.Errorf("invalid selection")
	}

	selected := instances[idx-1]
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(dimColor).PaddingLeft(2).Render(
		fmt.Sprintf("Connecting to %s (%s)...", selected.InstanceId, selected.InstanceName)))
	fmt.Println()

	tatClient, err := client.TAT()
	if err != nil {
		return err
	}

	session := &connectSession{
		instanceId: selected.InstanceId,
		cwd:        "/root",
		timeout:    60 * time.Second,
	}

	return session.runInteractive(tatClient, reader)
}

func workbenchExec(reader *bufio.Reader) error {
	if err := workbenchList(); err != nil {
		return err
	}

	promptStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	fmt.Print(promptStyle.Render("  Select instance (number): "))

	input, err := reader.ReadString('\n')
	if err != nil {
		return nil
	}

	idx, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return fmt.Errorf("invalid number")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	cvmClient, err := client.CVM()
	if err != nil {
		return err
	}

	instances, err := cvmClient.DescribeInstances(context.Background(), nil, "")
	if err != nil {
		return err
	}

	if idx < 1 || idx > len(instances) {
		return fmt.Errorf("invalid selection")
	}

	instanceId := instances[idx-1].InstanceId

	fmt.Print(promptStyle.Render("  Command to execute: "))
	cmdInput, err := reader.ReadString('\n')
	if err != nil {
		return nil
	}

	command := strings.TrimSpace(cmdInput)
	if command == "" {
		return fmt.Errorf("empty command")
	}

	tatClient, err := client.TAT()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(dimColor).PaddingLeft(2).Render("Executing..."))

	invocationId, err := tatClient.RunCommand(context.Background(), []string{instanceId}, command, "SHELL")
	if err != nil {
		return err
	}

	fmt.Println(lipgloss.NewStyle().Foreground(dimColor).PaddingLeft(2).Render("Invocation ID: " + invocationId))

	results, err := tatClient.WaitForCommand(context.Background(), invocationId, []string{instanceId}, 60*time.Second)
	if err != nil {
		return err
	}

	for _, r := range results {
		statusColor := successColor
		if r.Status != "SUCCESS" {
			statusColor = errorColor
		}
		fmt.Println()
		fmt.Printf("  %s  %s | exit: %d\n",
			lipgloss.NewStyle().Foreground(statusColor).Render("●"),
			lipgloss.NewStyle().Foreground(textColor).Render(r.Status),
			r.ExitCode)
		if r.Output != "" {
			fmt.Println()
			fmt.Println(r.Output)
		}
	}

	fmt.Println()
	return nil
}

func workbenchUpload(reader *bufio.Reader) error {
	if err := workbenchList(); err != nil {
		return err
	}

	promptStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	fmt.Print(promptStyle.Render("  Select instance (number): "))

	input, err := reader.ReadString('\n')
	if err != nil {
		return nil
	}

	idx, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return fmt.Errorf("invalid number")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	cvmClient, err := client.CVM()
	if err != nil {
		return err
	}

	instances, err := cvmClient.DescribeInstances(context.Background(), nil, "")
	if err != nil {
		return err
	}

	if idx < 1 || idx > len(instances) {
		return fmt.Errorf("invalid selection")
	}

	instanceId := instances[idx-1].InstanceId

	fmt.Print(promptStyle.Render("  Local file path: "))
	localInput, err := reader.ReadString('\n')
	if err != nil {
		return nil
	}
	localPath := strings.TrimSpace(localInput)

	fmt.Print(promptStyle.Render("  Remote path: "))
	remoteInput, err := reader.ReadString('\n')
	if err != nil {
		return nil
	}
	remotePath := strings.TrimSpace(remoteInput)

	tatClient, err := client.TAT()
	if err != nil {
		return err
	}

	onStatus := func(status string) {
		fmt.Fprintf(os.Stderr, "[%s] ", status)
	}

	return uploadFile(context.Background(), instanceId, localPath, remotePath, tatClient, onStatus)
}

func workbenchTasks(reader *bufio.Reader) error {
	promptStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	fmt.Print(promptStyle.Render("  Invocation ID: "))

	input, err := reader.ReadString('\n')
	if err != nil {
		return nil
	}

	invocationId := strings.TrimSpace(input)
	if invocationId == "" {
		return fmt.Errorf("empty invocation id")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	tatClient, err := client.TAT()
	if err != nil {
		return err
	}

	results, err := tatClient.DescribeInvocationTasks(context.Background(), invocationId, nil)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Foreground(dimColor).PaddingLeft(2).Render("No tasks found."))
		fmt.Println()
		return nil
	}

	fmt.Println()
	for _, r := range results {
		statusColor := successColor
		if r.Status != "SUCCESS" {
			statusColor = errorColor
		}
		fmt.Printf("  %s  %s | exit: %d\n",
			lipgloss.NewStyle().Foreground(statusColor).Render("●"),
			lipgloss.NewStyle().Foreground(textColor).Render(r.Status),
			r.ExitCode)
		if r.Output != "" {
			fmt.Println()
			fmt.Println(r.Output)
		}
		fmt.Println()
	}

	return nil
}
