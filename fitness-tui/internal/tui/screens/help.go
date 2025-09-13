package screens

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpItem struct {
	key         string
	description string
}

func (i helpItem) Title() string       { return i.key }
func (i helpItem) Description() string { return i.description }
func (i helpItem) FilterValue() string { return i.key }

type Help struct {
	list   list.Model
	width  int
	height int
}

func NewHelp() *Help {
	items := []list.Item{
		helpItem{"↑↓", "Navigate items"},
		helpItem{"Enter", "Select item"},
		helpItem{"s", "Sync activities"},
		helpItem{"a", "View activities"},
		helpItem{"c", "View charts"},
		helpItem{"q", "Return/Quit"},
		helpItem{"h/?", "Show this help"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("243"))

	l := list.New(items, delegate, 0, 0)
	l.Title = "Keyboard Shortcuts"
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true).
		MarginLeft(2)

	return &Help{list: l}
}

func (m *Help) Init() tea.Cmd {
	return nil
}

func (m *Help) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Help) View() string {
	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(m.list.View())
}
