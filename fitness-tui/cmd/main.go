package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/sstent/fitness-tui/internal/analysis"
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

	analyzeCmd := &cobra.Command{
		Use:   "analyze <activity-id>",
		Short: "Analyze a single activity with verbose logging",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			startTime := time.Now()
			logger := &garmin.CLILogger{}
			activityID := args[0]
			goal, _ := cmd.Flags().GetString("goal")

			if goal == "" {
				goal = "endurance" // default goal
			}

			logger.Infof("Starting analysis for activity %s with goal: %s", activityID, goal)

			cfg, err := config.Load()
			if err != nil {
				logger.Errorf("Config error: %v", err)
				os.Exit(1)
			}

			activityStorage := storage.NewActivityStorage(cfg.StoragePath)
			activity, err := activityStorage.Get(activityID)
			if err != nil {
				logger.Errorf("Activity load error: %v", err)
				os.Exit(1)
			}

			logger.Infof("Loaded activity: %s (%s)", activity.Name, activityID)
			logger.Infof("Activity Type: %s, Distance: %.2f km, Duration: %s",
				activity.Type, activity.Distance/1000, activity.FormattedDuration())

			params := analysis.PromptParams{
				Activity: activity,
				Goal:     goal,
				Config:   cfg,
			}
			orClient := analysis.NewOpenRouterClient(cfg)
			logger.Infof("Sending analysis request to OpenRouter using model: %s", cfg.OpenRouter.Model)

			ctx := context.Background()
			analysisResult, err := orClient.AnalyzeActivity(ctx, params)
			if err != nil {
				logger.Errorf("Analysis failed: %v", err)
				os.Exit(1)
			}

			duration := time.Since(startTime)
			logger.Infof("Analysis completed in %s!", duration.Round(time.Millisecond))

			fmt.Println("\n--- ANALYSIS RESULT ---")
			fmt.Println(analysisResult)
			fmt.Println("-----------------------")
		},
	}
	analyzeCmd.Flags().StringP("goal", "g", "", "Workout goal (e.g., endurance, intervals, recovery)")

	rootCmd.AddCommand(tuiCmd, syncCmd, analyzeCmd)
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

	app := tui.NewApp(activityStorage, garminClient, fileLogger, cfg)
	if err := app.Run(); err != nil {
		fmt.Printf("Application error: %v\n", err)
		os.Exit(1)
	}
}
