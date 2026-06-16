package cmd

import (
	"fmt"
	"os"
	"tcvm/internal/config"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Tencent Cloud credentials",
	Long:  `Interactive configuration for Tencent Cloud API credentials and COS settings.`,
	RunE:  runConfig,
}

var (
	configSecretId  string
	configSecretKey string
	configRegion    string
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().StringVar(&configSecretId, "secret-id", "", "Tencent Cloud SecretId")
	configCmd.Flags().StringVar(&configSecretKey, "secret-key", "", "Tencent Cloud SecretKey")
	configCmd.Flags().StringVar(&configRegion, "region", "", "Tencent Cloud Region (e.g., ap-guangzhou)")
}

func runConfig(cmd *cobra.Command, args []string) error {
	_ = config.InitConfig()

	if configSecretId != "" && configSecretKey != "" && configRegion != "" {
		return saveAndConfirm(configSecretId, configSecretKey, configRegion, config.AppConfig.CosBucket, config.AppConfig.CosRegion)
	}

	var (
		secretId  string
		secretKey string
		region    string
		cosBucket string
		cosRegion string
	)

	if config.AppConfig.SecretId != "" {
		secretId = config.AppConfig.SecretId
	}
	if config.AppConfig.SecretKey != "" {
		secretKey = config.AppConfig.SecretKey
	}
	if config.AppConfig.Region != "" {
		region = config.AppConfig.Region
	}
	if config.AppConfig.CosBucket != "" {
		cosBucket = config.AppConfig.CosBucket
	}
	if config.AppConfig.CosRegion != "" {
		cosRegion = config.AppConfig.CosRegion
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("SecretId").
				Description("Your Tencent Cloud API SecretId").
				Value(&secretId).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("secret-id is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("SecretKey").
				Description("Your Tencent Cloud API SecretKey").
				EchoMode(huh.EchoModePassword).
				Value(&secretKey).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("secret-key is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Region").
				Description("e.g. ap-guangzhou, ap-shanghai, ap-beijing, ap-hongkong, ap-singapore, na-siliconvalley...").
				Placeholder("ap-guangzhou").
				Value(&region).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("region is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("COS Bucket (optional)").
				Description("For large file transfers, e.g. mybucket-1250000000").
				Placeholder("").
				Value(&cosBucket),
			huh.NewInput().
				Title("COS Region (optional)").
				Description("e.g. ap-chengdu, ap-guangzhou").
				Placeholder("").
				Value(&cosRegion),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("configuration cancelled: %w", err)
	}

	return saveAndConfirm(secretId, secretKey, region, cosBucket, cosRegion)
}

func saveAndConfirm(secretId, secretKey, region, cosBucket, cosRegion string) error {
	if err := config.SaveConfig(secretId, secretKey, region, cosBucket, cosRegion); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Configuration saved to %s\n", config.ConfigPath())
	fmt.Fprintln(os.Stderr, "You can now use tcvm to manage your CVM instances.")
	return nil
}
