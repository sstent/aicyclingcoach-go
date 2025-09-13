package main

import (
	"fmt"
	"os"

	"github.com/sstent/fitness-tui/internal/config"
	"github.com/sstent/fitness-tui/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Using storage path: %s\n", cfg.StoragePath)

	app := tui.App{}
	if err := app.Run(); err != nil {
		fmt.Printf("Application error: %v\n", err)
		os.Exit(1)
	}
}
