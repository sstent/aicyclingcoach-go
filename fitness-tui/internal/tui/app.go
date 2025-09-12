package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	currentModel tea.Model
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}
	}
	return a, nil
}

func (a *App) View() string {
	return "AICyclingCoach-GO\n\nPress q to quit\n"
}

func (a *App) Run() error {
	p := tea.NewProgram(a)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run application: %w", err)
	}
	return nil
}
