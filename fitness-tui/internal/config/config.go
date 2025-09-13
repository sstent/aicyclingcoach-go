package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Garmin struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"garmin"`
	OpenRouter struct {
		APIKey string `mapstructure:"apikey"`
		Model  string `mapstructure:"model"`
	} `mapstructure:"openrouter"`
	StoragePath string `mapstructure:"storagepath"`
}

func Load() (*Config, error) {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".fitness-tui")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add search paths for config file
	viper.AddConfigPath(".")                                // Current directory
	viper.AddConfigPath(configDir)                          // ~/.fitness-tui/
	viper.AddConfigPath(filepath.Join(".", ".fitness-tui")) // ./.fitness-tui/

	setViperDefaults()

	// Read configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found - expected config.yaml in: %s", configDir)
		} else {
			return nil, fmt.Errorf("config read error: %w", err)
		}
	}

	// Create storage path atomically
	storagePath := viper.GetString("storagepath")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage path: %w", err)
	}

	cfg := new(Config)
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal error: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func setViperDefaults() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "." // Fallback to current directory
	}

	viper.SetDefault("storagepath", filepath.Join(home, ".fitness-tui"))
	viper.SetDefault("garmin.username", "")
	viper.SetDefault("garmin.password", "")
	viper.SetDefault("openrouter.apikey", "")
	viper.SetDefault("openrouter.model", "deepseek/deepseek-r1-0528")
}

func validateConfig(cfg *Config) error {
	switch {
	case cfg.Garmin.Username == "":
		return fmt.Errorf("garmin.username required")
	case cfg.Garmin.Password == "":
		return fmt.Errorf("garmin.password required")
	case cfg.OpenRouter.APIKey == "":
		return fmt.Errorf("openrouter.apikey required")
	}
	return nil
}
