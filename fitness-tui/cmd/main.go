package main

import (
	"fmt"
	"os"

	"github.com/sstent/fitness-tui/internal/config"
	"github.com/sstent/fitness-tui/internal/garmin"
	"github.com/sstent/fitness-tui/internal/storage"
	"github.com/sstent/fitness-tui/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize storage
	activityStorage := storage.NewActivityStorage(cfg.StoragePath)

	// Initialize Garmin client
	garminClient := garmin.NewClient(cfg.Garmin.Username, cfg.Garmin.Password, cfg.StoragePath)

	// Create and run the application
	app := tui.NewApp(activityStorage, garminClient)
	if err := app.Run(); err != nil {
		fmt.Printf("Application error: %v\n", err)
		os.Exit(1)
	}
}
