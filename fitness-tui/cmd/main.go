package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/sstent/fitness-tui/internal/config"
	"github.com/sstent/fitness-tui/internal/garmin"
	"github.com/sstent/fitness-tui/internal/storage"
	"github.com/sstent/fitness-tui/internal/tui"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "fitness-tui",
		Short: "Terminal-based fitness companion with AI analysis",
	}

	tuiCmd := &cobra.Command{
		Use:   "tui",
		Short: "Start the terminal user interface",
		Run: func(cmd *cobra.Command, args []string) {
			runTUI()
		},
	}

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync activities and files from Garmin Connect",
		Run: func(cmd *cobra.Command, args []string) {
			logger := &garmin.CLILogger{}
			logger.Infof("Starting sync process...")

			cfg, err := config.Load()
			if err != nil {
				logger.Errorf("Failed to load config: %v", err)
				os.Exit(1)
			}

			activityStorage := storage.NewActivityStorage(cfg.StoragePath)
			garminClient := garmin.NewClient(cfg.Garmin.Username, cfg.Garmin.Password, cfg.StoragePath)

			// Use the new Sync method that handles file downloads
			count, err := garminClient.Sync(context.Background(), activityStorage, logger)
			if err != nil {
				logger.Errorf("Sync failed: %v", err)
				os.Exit(1)
			}

			logger.Infof("Successfully synced %d activities with files", count)
		},
	}

	rootCmd.AddCommand(tuiCmd, syncCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runTUI() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize file logger
	logPath := filepath.Join(cfg.StoragePath, "fitness-tui.log")
	fileLogger, err := garmin.NewFileLogger(logPath)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer fileLogger.Close()

	activityStorage := storage.NewActivityStorage(cfg.StoragePath)
	garminClient := garmin.NewClient(cfg.Garmin.Username, cfg.Garmin.Password, cfg.StoragePath)

	app := tui.NewApp(activityStorage, garminClient, fileLogger)
	if err := app.Run(); err != nil {
		fmt.Printf("Application error: %v\n", err)
		os.Exit(1)
	}
}
