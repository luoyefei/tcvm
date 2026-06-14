package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	SecretId   string `mapstructure:"secret_id"`
	SecretKey  string `mapstructure:"secret_key"`
	Region     string `mapstructure:"region"`
	CosBucket  string `mapstructure:"cos_bucket"`
	CosRegion  string `mapstructure:"cos_region"`
}

var (
	AppConfig Config
	cfgFile   string
)

func InitConfig() error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to find home directory: %w", err)
		}
		configDir := filepath.Join(home, ".tcvm")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("unable to create config directory: %w", err)
		}
		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("TCVM")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return fmt.Errorf("unable to decode config: %w", err)
	}

	return nil
}

func SetConfigFile(file string) {
	cfgFile = file
}

func SaveConfig(secretId, secretKey, region, cosBucket, cosRegion string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".tcvm")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	viper.Set("secret_id", secretId)
	viper.Set("secret_key", secretKey)
	viper.Set("region", region)
	viper.Set("cos_bucket", cosBucket)
	viper.Set("cos_region", cosRegion)

	configPath := filepath.Join(configDir, "config.yaml")
	if err := viper.WriteConfigAs(configPath); err != nil {
		return err
	}

	AppConfig = Config{
		SecretId:  secretId,
		SecretKey: secretKey,
		Region:    region,
		CosBucket: cosBucket,
		CosRegion: cosRegion,
	}

	if err := os.Chmod(configPath, 0600); err != nil {
		return err
	}

	return nil
}

func IsConfigured() bool {
	return AppConfig.SecretId != "" && AppConfig.SecretKey != "" && AppConfig.Region != ""
}

func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tcvm", "config.yaml")
}
