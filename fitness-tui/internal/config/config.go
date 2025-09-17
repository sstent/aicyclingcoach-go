package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Garmin struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	} `mapstructure:"garmin"`
	OpenRouter struct {
		APIKey  string        `mapstructure:"apikey"`
		Model   string        `mapstructure:"model"`
		BaseURL string        `mapstructure:"base_url"`
		Timeout time.Duration `mapstructure:"timeout"`
	} `mapstructure:"openrouter"`
	Metrics struct {
		Cycling struct {
			Power     bool `mapstructure:"power"`
			Cadence   bool `mapstructure:"cadence"`
			Elevation bool `mapstructure:"elevation"`
		} `mapstructure:"cycling"`
		Running struct {
			Cadence             bool `mapstructure:"cadence"`
			VerticalOscillation bool `mapstructure:"vertical_oscillation"`
		} `mapstructure:"running"`
		Hiking struct {
			Temperature bool `mapstructure:"temperature"`
			AscentRate  bool `mapstructure:"ascent_rate"`
		} `mapstructure:"hiking"`
		Generic struct {
			HeartRate bool `mapstructure:"heart_rate"`
			Speed     bool `mapstructure:"speed"`
			Duration  bool `mapstructure:"duration"`
		} `mapstructure:"generic"`
		DetailLevel int `mapstructure:"detail_level"`
	} `mapstructure:"metrics"`
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
	viper.SetDefault("openrouter.base_url", "https://openrouter.ai/api/v1")
	viper.SetDefault("openrouter.timeout", 30*time.Second)
	viper.SetDefault("openrouter.base_url", "https://openrouter.ai/api/v1")
	viper.SetDefault("openrouter.timeout", "30s")
	viper.SetDefault("openrouter.base_url", "https://openrouter.ai/api/v1")
	viper.SetDefault("openrouter.timeout", 30*time.Second)

	// Metrics defaults
	viper.SetDefault("metrics.cycling.power", true)
	viper.SetDefault("metrics.cycling.cadence", true)
	viper.SetDefault("metrics.cycling.elevation", true)
	viper.SetDefault("metrics.running.cadence", true)
	viper.SetDefault("metrics.running.vertical_oscillation", true)
	viper.SetDefault("metrics.hiking.temperature", true)
	viper.SetDefault("metrics.hiking.ascent_rate", true)
	viper.SetDefault("metrics.generic.heart_rate", true)
	viper.SetDefault("metrics.generic.speed", true)
	viper.SetDefault("metrics.generic.duration", true)
	viper.SetDefault("metrics.detail_level", 2) // Default to medium detail
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
