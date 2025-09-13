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
	activityList    *screens.ActivityList // Persistent activity list
	width           int                   // Track window width
	height          int                   // Track window height
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
		activityList:    activityList, // Store persistent reference
	}
}

func (a *App) Init() tea.Cmd {
	return a.currentModel.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Store window size and forward to current model
		a.width = msg.Width
		a.height = msg.Height
		updatedModel, cmd := a.currentModel.Update(msg)
		a.currentModel = updatedModel
		return a, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Only quit if we're at the top level (activity list)
			if _, ok := a.currentModel.(*screens.ActivityList); ok {
				return a, tea.Quit
			}
		}
	case screens.ActivitySelectedMsg:
		a.logger.Debugf("App.Update() - Received ActivitySelectedMsg for: %s", msg.Activity.Name)
		// For now, use empty analysis - we'll implement analysis caching later
		detail := screens.NewActivityDetail(msg.Activity, "", a.logger)
		a.currentModel = detail
		return a, detail.Init()
	case screens.BackToListMsg:
		a.logger.Debugf("App.Update() - Received BackToListMsg")
		// Return to existing activity list instead of creating new
		a.currentModel = a.activityList
		// Send current window size to ensure proper rendering
		return a, func() tea.Msg {
			return tea.WindowSizeMsg{Width: a.width, Height: a.height}
		}
	}

	// Delegate to the current model
	updatedModel, cmd := a.currentModel.Update(msg)
	a.currentModel = updatedModel

	// Update activity list reference if needed
	if activityList, ok := updatedModel.(*screens.ActivityList); ok {
		a.activityList = activityList
	}

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
