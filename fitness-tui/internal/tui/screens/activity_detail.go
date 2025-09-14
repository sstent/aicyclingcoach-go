// internal/tui/screens/activity_detail.go
package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/sstent/fitness-tui/internal/garmin"
	"github.com/sstent/fitness-tui/internal/tui/components"
	"github.com/sstent/fitness-tui/internal/tui/layout"
	"github.com/sstent/fitness-tui/internal/tui/models"
)

type BackToListMsg struct{}

type ActivityDetail struct {
	activity       *models.Activity
	analysis       string
	viewport       viewport.Model
	layout         *layout.Layout
	hrChart        *components.Chart
	elevationChart *components.Chart
	logger         garmin.Logger
	ready          bool
	currentTab     int // 0: Overview, 1: Charts, 2: Analysis
	tabNames       []string
}

func NewActivityDetail(activity *models.Activity, analysis string, logger garmin.Logger) *ActivityDetail {
	if logger == nil {
		logger = &garmin.NoopLogger{}
	}

	vp := viewport.New(80, 20)
	ad := &ActivityDetail{
		activity:       activity,
		analysis:       analysis,
		viewport:       vp,
		layout:         layout.NewLayout(80, 24),
		logger:         logger,
		hrChart:        components.NewChart(activity.Metrics.HeartRateData, 40, 4, "Heart Rate (bpm)"),
		elevationChart: components.NewChart(activity.Metrics.ElevationData, 40, 4, "Elevation (m)"),
		tabNames:       []string{"Overview", "Charts", "Analysis"},
	}
	ad.setContent()
	return ad
}

func (m *ActivityDetail) Init() tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{Width: 80, Height: 24}
	}
}

func (m *ActivityDetail) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.layout = layout.NewLayout(msg.Width, msg.Height)

		// Calculate viewport height dynamically
		headerHeight := 3
		tabHeight := 3
		navHeight := 2
		helpHeight := 1
		padding := 2
		viewportHeight := msg.Height - headerHeight - tabHeight - navHeight - helpHeight - padding

		m.viewport = viewport.New(msg.Width-4, viewportHeight)
		m.ready = true
		m.setContent()
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "b", "q":
			return m, func() tea.Msg { return BackToListMsg{} }
		case "tab", "right", "l":
			m.currentTab = (m.currentTab + 1) % len(m.tabNames)
			m.setContent()
		case "shift+tab", "left", "h":
			m.currentTab = (m.currentTab - 1 + len(m.tabNames)) % len(m.tabNames)
			m.setContent()
		case "1":
			m.currentTab = 0
			m.setContent()
		case "2":
			m.currentTab = 1
			m.setContent()
		case "3":
			m.currentTab = 2
			m.setContent()
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *ActivityDetail) View() string {
	if !m.ready {
		return "Loading activity details..."
	}

	var content strings.Builder

	// Header with activity name
	breadcrumb := "Home > Activities > " + m.activity.Name
	content.WriteString(m.layout.HeaderPanel(m.activity.Name, breadcrumb))

	// Tab navigation
	content.WriteString(m.renderTabNavigation())

	// Main content area - use remaining height
	content.WriteString(m.viewport.View())

	// Navigation bar
	navItems := []layout.NavItem{
		{Label: "Overview", Key: "1"},
		{Label: "Charts", Key: "2"},
		{Label: "Analysis", Key: "3"},
		{Label: "Back", Key: "esc"},
	}
	content.WriteString(m.layout.NavigationBar(navItems, m.currentTab))

	// Help text
	helpText := "1-3 switch tabs • ←→ navigate tabs • esc back • q quit"
	content.WriteString(m.layout.HelpText(helpText))

	return m.layout.MainContainer().
		Height(m.layout.Height). // Use full height
		Render(content.String())
}

func (m *ActivityDetail) renderTabNavigation() string {
	var tabs []string
	tabWidth := (m.layout.Width - 8) / len(m.tabNames)

	for i, tabName := range m.tabNames {
		var tabStyle lipgloss.Style
		if i == m.currentTab {
			tabStyle = lipgloss.NewStyle().
				Width(tabWidth).
				Height(3).
				Background(layout.CardBG).
				Foreground(layout.PrimaryOrange).
				Bold(true).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(layout.PrimaryOrange).
				Padding(1).
				Align(lipgloss.Center)
		} else {
			tabStyle = lipgloss.NewStyle().
				Width(tabWidth).
				Height(3).
				Background(layout.LightBG).
				Foreground(layout.MutedText).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(layout.MutedText).
				Padding(1).
				Align(lipgloss.Center)
		}
		tabs = append(tabs, tabStyle.Render(tabName))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...) + "\n"
}

func (m *ActivityDetail) setContent() {
	var content strings.Builder

	if m.activity == nil {
		content.WriteString("Activity data is nil!")
		m.viewport.SetContent(content.String())
		return
	}

	switch m.currentTab {
	case 0: // Overview
		content.WriteString(m.renderOverviewTab())
	case 1: // Charts
		content.WriteString(m.renderChartsTab())
	case 2: // Analysis
		content.WriteString(m.renderAnalysisTab())
	}

	m.viewport.SetContent(content.String())
}

func (m *ActivityDetail) renderOverviewTab() string {
	var content strings.Builder

	// Activity stats cards
	content.WriteString(m.renderStatsCards())
	content.WriteString("\n\n")

	// Two-column layout for detailed metrics
	leftContent := m.renderBasicMetrics()
	rightContent := m.renderPerformanceMetrics()

	content.WriteString(m.layout.TwoColumnLayout(leftContent, rightContent, m.layout.Width/2))

	return content.String()
}

func (m *ActivityDetail) renderStatsCards() string {
	cardWidth := (m.layout.Width - 16) / 4

	cards := []string{
		m.layout.StatCard("Duration", m.activity.FormattedDuration(), layout.PrimaryBlue, cardWidth),
		m.layout.StatCard("Distance", m.activity.FormattedDistance(), layout.PrimaryGreen, cardWidth),
		m.layout.StatCard("Avg Pace", m.activity.FormattedPace(), layout.PrimaryOrange, cardWidth),
		m.layout.StatCard("Calories", fmt.Sprintf("%d", m.activity.Calories), layout.PrimaryPink, cardWidth),
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}

func (m *ActivityDetail) renderBasicMetrics() string {
	var content strings.Builder

	content.WriteString(lipgloss.NewStyle().
		Foreground(layout.PrimaryPurple).
		Bold(true).
		MarginBottom(1).
		Render("Activity Details"))
	content.WriteString("\n\n")

	metrics := []struct {
		label string
		value string
		color lipgloss.Color
	}{
		{"Date", m.activity.Date.Format("Monday, January 2, 2006"), layout.LightText},
		{"Type", strings.Title(m.activity.Type), layout.PrimaryBlue},
		{"Duration", m.activity.FormattedDuration(), layout.PrimaryGreen},
		{"Distance", m.activity.FormattedDistance(), layout.PrimaryOrange},
		{"Calories", fmt.Sprintf("%d kcal", m.activity.Calories), layout.PrimaryPink},
	}

	for _, metric := range metrics {
		content.WriteString(lipgloss.NewStyle().
			Foreground(layout.MutedText).
			Width(15).
			Render(metric.label + ":"))
		content.WriteString(" ")
		content.WriteString(lipgloss.NewStyle().
			Foreground(metric.color).
			Bold(true).
			Render(metric.value))
		content.WriteString("\n")
	}

	return content.String()
}

func (m *ActivityDetail) renderPerformanceMetrics() string {
	var content strings.Builder

	content.WriteString(lipgloss.NewStyle().
		Foreground(layout.PrimaryYellow).
		Bold(true).
		MarginBottom(1).
		Render("Performance Metrics"))
	content.WriteString("\n\n")

	metrics := []struct {
		label string
		value string
		color lipgloss.Color
	}{
		{"Avg Heart Rate", fmt.Sprintf("%d bpm", m.activity.Metrics.AvgHeartRate), layout.PrimaryPink},
		{"Max Heart Rate", fmt.Sprintf("%d bpm", m.activity.Metrics.MaxHeartRate), layout.PrimaryPink},
		{"Avg Speed", fmt.Sprintf("%.1f km/h", m.activity.Metrics.AvgSpeed), layout.PrimaryBlue},
		{"Elevation Gain", fmt.Sprintf("%.0f m", m.activity.Metrics.ElevationGain), layout.PrimaryGreen},
		{"Training Stress", fmt.Sprintf("%.1f TSS", m.activity.Metrics.TrainingStress), layout.PrimaryOrange},
		{"Recovery Time", fmt.Sprintf("%d hours", m.activity.Metrics.RecoveryTime), layout.PrimaryPurple},
		{"Intensity Factor", fmt.Sprintf("%.2f", m.activity.Metrics.IntensityFactor), layout.PrimaryYellow},
	}

	for _, metric := range metrics {
		content.WriteString(lipgloss.NewStyle().
			Foreground(layout.MutedText).
			Width(18).
			Render(metric.label + ":"))
		content.WriteString(" ")
		content.WriteString(lipgloss.NewStyle().
			Foreground(metric.color).
			Bold(true).
			Render(metric.value))
		content.WriteString("\n")
	}

	return content.String()
}

func (m *ActivityDetail) renderChartsTab() string {
	var content strings.Builder

	content.WriteString(lipgloss.NewStyle().
		Foreground(layout.PrimaryBlue).
		Bold(true).
		MarginBottom(2).
		Render("Performance Charts"))
	content.WriteString("\n\n")

	if len(m.activity.Metrics.HeartRateData) == 0 && len(m.activity.Metrics.ElevationData) == 0 {
		content.WriteString(lipgloss.NewStyle().
			Foreground(layout.MutedText).
			Align(lipgloss.Center).
			Width(m.layout.Width - 8).
			Render("No chart data available for this activity"))
		return content.String()
	}

	// Update chart dimensions for full-width display
	chartWidth := m.layout.Width - 12
	chartHeight := (m.layout.Height - 15) / 2

	m.hrChart.Width = chartWidth
	m.hrChart.Height = chartHeight
	m.elevationChart.Width = chartWidth
	m.elevationChart.Height = chartHeight

	if len(m.activity.Metrics.HeartRateData) > 0 {
		content.WriteString(m.layout.ChartPanel("Heart Rate", m.hrChart.View(), layout.PrimaryPink))
		content.WriteString("\n")
	}

	if len(m.activity.Metrics.ElevationData) > 0 {
		content.WriteString(m.layout.ChartPanel("Elevation", m.elevationChart.View(), layout.PrimaryGreen))
		content.WriteString("\n")
	}

	// Chart legend/info
	content.WriteString(lipgloss.NewStyle().
		Foreground(layout.MutedText).
		Italic(true).
		MarginTop(1).
		Render("Charts show real-time data throughout the activity duration"))

	return content.String()
}

func (m *ActivityDetail) renderAnalysisTab() string {
	var content strings.Builder

	content.WriteString(lipgloss.NewStyle().
		Foreground(layout.PrimaryGreen).
		Bold(true).
		MarginBottom(2).
		Render("AI Analysis"))
	content.WriteString("\n\n")

	if m.analysis == "" {
		content.WriteString(lipgloss.NewStyle().
			Foreground(layout.MutedText).
			Align(lipgloss.Center).
			Width(m.layout.Width - 8).
			Render("No AI analysis available for this activity\nAnalysis will be generated automatically in future updates"))
		return content.String()
	}

	// Parse and format analysis sections
	sections := strings.Split(m.analysis, "## ")
	for i, section := range sections {
		if strings.TrimSpace(section) == "" {
			continue
		}

		// Split section into title and content
		parts := strings.SplitN(section, "\n", 2)
		if len(parts) < 2 {
			content.WriteString(lipgloss.NewStyle().
				Foreground(layout.LightText).
				Render(section))
			content.WriteString("\n\n")
			continue
		}

		title := strings.TrimSpace(parts[0])
		body := strings.TrimSpace(parts[1])

		// Use different colors for different sections
		colors := []lipgloss.Color{
			layout.PrimaryBlue,
			layout.PrimaryGreen,
			layout.PrimaryOrange,
			layout.PrimaryPink,
			layout.PrimaryPurple,
		}
		sectionColor := colors[i%len(colors)]

		// Render section title
		content.WriteString(lipgloss.NewStyle().
			Foreground(sectionColor).
			Bold(true).
			MarginBottom(1).
			Render("🔍 " + title))
		content.WriteString("\n")

		// Format content with bullet points
		lines := strings.Split(body, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "- ") {
				content.WriteString(lipgloss.NewStyle().
					Foreground(layout.LightText).
					Render("  • " + strings.TrimPrefix(line, "- ")))
			} else if strings.TrimSpace(line) != "" {
				content.WriteString(lipgloss.NewStyle().
					Foreground(layout.LightText).
					Render(line))
			}
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	return content.String()
}
