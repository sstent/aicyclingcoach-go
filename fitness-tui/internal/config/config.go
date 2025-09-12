package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Garmin struct {
		Username string
		Password string
	}
	OpenRouter struct {
		APIKey string
		Model  string
	}
	StoragePath string
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Join(home, ".fitness-tui"))
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set defaults if not configured
	if cfg.StoragePath == "" {
		cfg.StoragePath = filepath.Join(home, ".fitness-tui")
	}
	if cfg.OpenRouter.Model == "" {
		cfg.OpenRouter.Model = "deepseek/deepseek-r1-0528"
	}

	return &cfg, nil
}
