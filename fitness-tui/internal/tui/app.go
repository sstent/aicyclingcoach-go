// internal/tui/app.go
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
	screenStack     []tea.Model           // Screen navigation stack
}

func NewApp(activityStorage *storage.ActivityStorage, garminClient *garmin.Client, logger garmin.Logger) *App {
	if logger == nil {
		logger = &garmin.NoopLogger{}
	}

	// Initialize with a placeholder screen - actual size will be set by WindowSizeMsg
	activityList := screens.NewActivityList(activityStorage, garminClient)

	app := &App{
		currentModel:    activityList,
		activityStorage: activityStorage,
		garminClient:    garminClient,
		logger:          logger,
		activityList:    activityList, // Store persistent reference
		screenStack:     []tea.Model{activityList},
		width:           80, // Default width
		height:          24, // Default height
	}
	return app
}

func (a *App) Init() tea.Cmd {
	return a.currentModel.Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Only update if size actually changed
		if a.width != msg.Width || a.height != msg.Height {
			a.width = msg.Width
			a.height = msg.Height
			updatedModel, cmd := a.currentModel.Update(msg)
			a.currentModel = updatedModel
			a.updateStackTop(updatedModel)
			return a, cmd
		}
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			// Force quit on Ctrl+C
			return a, tea.Quit
		case "q":
			// Handle quit - go back in stack or quit if at root
			if len(a.screenStack) <= 1 {
				// At root level, quit
				if _, ok := a.currentModel.(*screens.ActivityList); ok {
					return a, tea.Quit
				}
			} else {
				// Go back to previous screen
				return a, a.goBack()
			}
		}

	case screens.ActivitySelectedMsg:
		a.logger.Debugf("App.Update() - Received ActivitySelectedMsg for: %s", msg.Activity.Name)
		// Create new activity detail screen
		detail := screens.NewActivityDetail(msg.Activity, "", a.logger)
		a.pushScreen(detail)
		return a, detail.Init()

	case screens.BackToListMsg:
		a.logger.Debugf("App.Update() - Received BackToListMsg")
		return a, a.goBack()
	}

	// Delegate to the current model
	updatedModel, cmd := a.currentModel.Update(msg)
	a.currentModel = updatedModel
	a.updateStackTop(updatedModel)

	return a, cmd
}

func (a *App) View() string {
	return a.currentModel.View()
}

func (a *App) Run() error {
	// Use alt screen for better TUI experience
	p := tea.NewProgram(a, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run application: %w", err)
	}
	return nil
}

// pushScreen adds a new screen to the stack and makes it current
func (a *App) pushScreen(model tea.Model) {
	a.screenStack = append(a.screenStack, model)
	a.currentModel = model
}

// goBack removes the current screen from stack and returns to previous
func (a *App) goBack() tea.Cmd {
	if len(a.screenStack) <= 1 {
		// Already at root, can't go back further
		return nil
	}

	// Remove current screen
	a.screenStack = a.screenStack[:len(a.screenStack)-1]

	// Set previous screen as current
	a.currentModel = a.screenStack[len(a.screenStack)-1]

	// Update the model with current window size
	var cmd tea.Cmd
	a.currentModel, cmd = a.currentModel.Update(tea.WindowSizeMsg{Width: a.width, Height: a.height})
	a.updateStackTop(a.currentModel)
	return cmd
}

// updateStackTop updates the top of the stack with the current model
func (a *App) updateStackTop(model tea.Model) {
	if len(a.screenStack) > 0 {
		a.screenStack[len(a.screenStack)-1] = model
	}

	// Update activity list reference if needed
	if activityList, ok := model.(*screens.ActivityList); ok {
		a.activityList = activityList
	}
}
