package components

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
)

type Spinner struct {
	spinner  spinner.Model
	message  string
	quitting bool
}

func NewSpinner(message string) Spinner {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle
	return Spinner{
		spinner: s,
		message: message,
	}
}

func (s Spinner) Init() tea.Cmd {
	return s.spinner.Tick
}

func (s Spinner) Update(msg tea.Msg) (Spinner, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			s.quitting = true
			return s, tea.Quit
		default:
			return s, nil
		}
	default:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd
	}
}

func (s Spinner) View() string {
	str := lipgloss.JoinHorizontal(lipgloss.Center,
		s.spinner.View(),
		" "+s.message,
	)
	if s.quitting {
		return str + "\n"
	}
	return str
}

func (s Spinner) SetMessage(msg string) Spinner {
	s.message = msg
	return s
}
