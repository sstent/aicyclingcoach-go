package main

import (
	"context"
	"fmt"
	"os"

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
		Short: "Sync activities from Garmin Connect",
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

			// Authenticate
			if err := garminClient.Connect(logger); err != nil {
				logger.Errorf("Authentication failed: %v", err)
				os.Exit(1)
			}

			// Get activities
			activities, err := garminClient.GetActivities(context.Background(), 50, logger)
			if err != nil {
				logger.Errorf("Failed to fetch activities: %v", err)
				os.Exit(1)
			}

			// Save activities
			for i, activity := range activities {
				if err := activityStorage.Save(activity); err != nil {
					logger.Errorf("Failed to save activity %s: %v", activity.ID, err)
				} else {
					logger.Infof("[%d/%d] Saved activity: %s", i+1, len(activities), activity.Name)
				}
			}

			logger.Infof("Successfully synced %d activities", len(activities))
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

	activityStorage := storage.NewActivityStorage(cfg.StoragePath)
	garminClient := garmin.NewClient(cfg.Garmin.Username, cfg.Garmin.Password, cfg.StoragePath)

	app := tui.NewApp(activityStorage, garminClient)
	if err := app.Run(); err != nil {
		fmt.Printf("Application error: %v\n", err)
		os.Exit(1)
	}
}
