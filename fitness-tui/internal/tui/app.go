package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sstent/fitness-tui/internal/garmin"
	"github.com/sstent/fitness-tui/internal/storage"
	"github.com/sstent/fitness-tui/internal/tui/screens"
)

type App struct {
	currentModel tea.Model
}

func NewApp(activityStorage *storage.ActivityStorage, garminClient *garmin.Client) *App {
	// Initialize with the activity list screen as the default
	activityList := screens.NewActivityList(activityStorage, garminClient)

	return &App{
		currentModel: activityList,
	}
}

func (a *App) Init() tea.Cmd {
	return a.currentModel.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}
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
