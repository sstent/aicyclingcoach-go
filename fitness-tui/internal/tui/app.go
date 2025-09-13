package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sstent/fitness-tui/internal/garmin"
	"github.com/sstent/fitness-tui/internal/storage"
	"github.com/sstent/fitness-tui/internal/tui/screens"
)

type App struct {
	currentModel    tea.Model
	activityStorage *storage.ActivityStorage
	garminClient    *garmin.Client
	logger          garmin.Logger
}

func NewApp(activityStorage *storage.ActivityStorage, garminClient *garmin.Client, logger garmin.Logger) *App {
	if logger == nil {
		logger = &garmin.NoopLogger{}
	}

	// Initialize with the activity list screen as the default
	activityList := screens.NewActivityList(activityStorage, garminClient)

	return &App{
		currentModel:    activityList,
		activityStorage: activityStorage,
		garminClient:    garminClient,
		logger:          logger,
	}
}

func (a *App) Init() tea.Cmd {
	return a.currentModel.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Forward window size to current model
		updatedModel, cmd := a.currentModel.Update(msg)
		a.currentModel = updatedModel
		return a, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}
	case screens.ActivitySelectedMsg:
		a.logger.Debugf("App.Update() - Received ActivitySelectedMsg for: %s", msg.Activity.Name)
		// For now, use empty analysis - we'll implement analysis caching later
		detail := screens.NewActivityDetail(msg.Activity, "", a.logger)
		a.currentModel = detail
		return a, detail.Init()
	case screens.BackToListMsg:
		a.logger.Debugf("App.Update() - Received BackToListMsg")
		// Re-initialize the activity list when navigating back
		activityList := screens.NewActivityList(a.activityStorage, a.garminClient)
		a.currentModel = activityList
		return a, activityList.Init()
	}

	// Delegate to the current model
	updatedModel, cmd := a.currentModel.Update(msg)
	a.currentModel = updatedModel
	return a, cmd
}

func (a *App) View() string {
	return a.currentModel.View()
}

func (a *App) Run() error {
	p := tea.NewProgram(a)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run application: %w", err)
	}
	return nil
}
