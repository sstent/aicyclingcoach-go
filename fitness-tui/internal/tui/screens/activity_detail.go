package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/sstent/fitness-tui/internal/tui/components"
	"github.com/sstent/fitness-tui/internal/tui/models"
)

type ActivityDetail struct {
	activity       *models.Activity
	analysis       string
	viewport       viewport.Model
	width          int
	height         int
	styles         *Styles
	hrChart        *components.Chart
	elevationChart *components.Chart
}

type Styles struct {
	Title     lipgloss.Style
	Subtitle  lipgloss.Style
	StatName  lipgloss.Style
	StatValue lipgloss.Style
	Analysis  lipgloss.Style
	Viewport  lipgloss.Style
}

func NewActivityDetail(activity *models.Activity, analysis string) *ActivityDetail {
	styles := &Styles{
		Title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")),
		Subtitle:  lipgloss.NewStyle().Foreground(lipgloss.Color("8")).MarginTop(1),
		StatName:  lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		StatValue: lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
		Analysis:  lipgloss.NewStyle().MarginTop(2),
		Viewport:  lipgloss.NewStyle().Padding(0, 2),
	}

	vp := viewport.New(0, 0)
	return &ActivityDetail{
		activity:       activity,
		analysis:       analysis,
		viewport:       vp,
		styles:         styles,
		hrChart:        components.NewChart(activity.Metrics.HeartRateData, 40, 4, "Heart Rate (bpm)"),
		elevationChart: components.NewChart(activity.Metrics.ElevationData, 40, 4, "Elevation (m)"),
	}
}

func (m *ActivityDetail) Init() tea.Cmd {
	return nil
}

func (m *ActivityDetail) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4
		chartWidth := msg.Width/2 - 4
		m.hrChart.Width = chartWidth
		m.elevationChart.Width = chartWidth
		m.setContent()
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *ActivityDetail) View() string {
	return m.viewport.View()
}

func (m *ActivityDetail) setContent() {
	var content strings.Builder

	// Activity Details
	content.WriteString(m.styles.Title.Render(m.activity.Name))
	content.WriteString("\n\n")

	content.WriteString(m.styles.Subtitle.Render("Activity Details"))
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("%s  %s\n",
		m.styles.StatName.Render("Date:"),
		m.styles.StatValue.Render(m.activity.Date.Format("2006-01-02")),
	))
	content.WriteString(fmt.Sprintf("%s  %s\n",
		m.styles.StatName.Render("Duration:"),
		m.styles.StatValue.Render(m.activity.FormattedDuration()),
	))
	content.WriteString(fmt.Sprintf("%s  %s\n",
		m.styles.StatName.Render("Distance:"),
		m.styles.StatValue.Render(m.activity.FormattedDistance()),
	))

	// Charts Section
	content.WriteString(m.styles.Subtitle.Render("Performance Charts"))
	content.WriteString("\n")
	charts := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.hrChart.View(),
		lipgloss.NewStyle().Width(2).Render(" "),
		m.elevationChart.View(),
	)
	content.WriteString(charts)
	content.WriteString("\n\n")

	// Analysis Section
	if m.analysis != "" {
		content.WriteString(m.styles.Analysis.Render(
			m.styles.Subtitle.Render("AI Analysis"),
		))
		content.WriteString("\n")
		content.WriteString(m.styles.StatValue.Render(m.analysis))
	}

	m.viewport.SetContent(m.styles.Viewport.Render(content.String()))
}
