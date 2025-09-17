// internal/tui/screens/activity_detail.go
package screens

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/sstent/fitness-tui/internal/analysis"
	"github.com/sstent/fitness-tui/internal/config"
	"github.com/sstent/fitness-tui/internal/garmin"
	"github.com/sstent/fitness-tui/internal/storage"
	"github.com/sstent/fitness-tui/internal/tui/components"
	"github.com/sstent/fitness-tui/internal/tui/models"
	"github.com/sstent/fitness-tui/internal/tui/styles"
)

type BackToListMsg struct{}
type AnalysisCompleteMsg struct {
	Analysis string
}

type AnalysisFailedMsg struct {
	Error error
}

type AnalysisProgressMsg struct {
	Progress string
}

type ActivityDetail struct {
	activity         *models.Activity
	analysis         string
	viewport         viewport.Model
	hrChart          *components.Chart
	powerChart       *components.Chart
	elevationChart   *components.Chart
	logger           garmin.Logger
	config           *config.Config
	styles           *styles.Styles
	ready            bool
	currentTab       int // 0: Overview, 1: Charts, 2: Analysis
	tabNames         []string
	generating       bool
	analysisSpinner  spinner.Model
	analysisProgress string
	lastError        error // Store the last analysis error
}

func NewActivityDetail(activity *models.Activity, analysis string, config *config.Config, logger garmin.Logger) *ActivityDetail {
	st := styles.NewStyles()
	if logger == nil {
		logger = &garmin.NoopLogger{}
	}

	vp := viewport.New(80, 20)
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	ad := &ActivityDetail{
		activity: activity,
		analysis: analysis,
		viewport: vp,
		logger:   logger,
		config:   config,
		styles:   st,
		hrChart: components.NewChart(
			activity.Metrics.HeartRateData,
			"Heart Rate",
			"bpm",
			40,
			4,
			lipgloss.Color("#FF0000"),
		),
		powerChart: components.NewChart(
			activity.Metrics.PowerData,
			"Power",
			"watts",
			40,
			4,
			lipgloss.Color("#00FF00"),
		),
		elevationChart: components.NewChart(
			activity.Metrics.ElevationData,
			"Elevation",
			"m",
			40,
			4,
			lipgloss.Color("#0000FF"),
		),
		tabNames:         []string{"Overview", "Charts", "Analysis"},
		analysisSpinner:  s,
		analysisProgress: "Ready to analyze",
	}
	ad.setContent()
	return ad
}

func (m *ActivityDetail) Init() tea.Cmd {
	return tea.Batch(
		m.analysisSpinner.Tick,
		func() tea.Msg {
			return tea.WindowSizeMsg{Width: 80, Height: 24}
		},
	)
}

func (m *ActivityDetail) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:

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
		case "a": // Trigger analysis (uses cached if available)
			if m.currentTab == 2 && !m.generating {
				m.generating = true
				m.analysisProgress = "Checking cache..."
				m.lastError = nil // Clear previous error
				cmds = append(cmds, tea.Batch(
					m.analysisSpinner.Tick,
					m.generateAnalysisCmd(false),
				))
			}
		case "A": // Force re-analyze (skip cache)
			if m.currentTab == 2 && !m.generating {
				m.generating = true
				m.analysisProgress = "Forcing new analysis..."
				m.lastError = nil // Clear previous error
				cmds = append(cmds, tea.Batch(
					m.analysisSpinner.Tick,
					m.generateAnalysisCmd(true),
				))
			}
		case "r": // Refresh or retry
			if m.currentTab == 2 {
				if !m.generating {
					if m.analysis == "" && m.lastError != nil {
						// Retry failed analysis
						m.generating = true
						m.analysisProgress = "Retrying analysis..."
						m.lastError = nil // Clear previous error
						cmds = append(cmds, tea.Batch(
							m.analysisSpinner.Tick,
							m.generateAnalysisCmd(false),
						))
					} else {
						// Refresh existing analysis
						m.analysis = ""
						m.analysisProgress = "Refreshing analysis..."
						m.generating = true
						m.lastError = nil // Clear previous error
						cmds = append(cmds, tea.Batch(
							m.analysisSpinner.Tick,
							m.generateAnalysisCmd(false),
						))
					}
				}
			}
		}
	case AnalysisCompleteMsg:
		m.generating = false
		if msg.Analysis != "" {
			m.analysis = msg.Analysis
		}
		m.analysisProgress = "Analysis complete"
		m.setContent()
	case AnalysisFailedMsg:
		m.generating = false
		m.lastError = msg.Error
		m.analysisProgress = "Analysis failed"
		m.setContent()
	case AnalysisProgressMsg:
		m.analysisProgress = msg.Progress
		m.setContent()
	}

	// Update spinner if generating
	if m.generating {
		m.analysisSpinner, cmd = m.analysisSpinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *ActivityDetail) View() string {
	if !m.ready {
		return "Loading activity details..."
	}

	var content strings.Builder

	// Header with activity name
	header := m.styles.HeaderPanel.Render(m.activity.Name)
	content.WriteString(header)

	// Tab navigation
	content.WriteString(m.renderTabNavigation())

	// Main content area - use remaining height
	content.WriteString(m.viewport.View())

	// Navigation bar
	navItems := []styles.NavItem{
		{Label: "Overview", Key: "1"},
		{Label: "Charts", Key: "2"},
		{Label: "Analysis", Key: "3"},
		{Label: "Back", Key: "esc"},
	}
	content.WriteString(m.styles.NavigationBar(navItems, m.currentTab))

	// Help text
	helpText := "1-3 switch tabs • ←→ navigate tabs • esc back"
	if m.currentTab == 2 {
		helpText += " • a: analyze • r: refresh/retry"
	}
	helpText += " • q quit"
	content.WriteString(m.styles.HelpText.Render(helpText))

	return m.styles.MainContainer.
		Render(content.String())
}

func (m *ActivityDetail) renderTabNavigation() string {
	var tabs []string
	tabWidth := (m.styles.Dimensions.Width - 8) / len(m.tabNames)

	for i, tabName := range m.tabNames {
		var tabStyle lipgloss.Style
		if i == m.currentTab {
			tabStyle = lipgloss.NewStyle().
				Width(tabWidth).
				Height(3).
				Background(m.styles.CardBG).
				Foreground(m.styles.PrimaryOrange).
				BorderForeground(m.styles.PrimaryOrange).
				Bold(true).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(m.styles.PrimaryOrange).
				Padding(1).
				Align(lipgloss.Center)
		} else {
			tabStyle = lipgloss.NewStyle().
				Width(tabWidth).
				Height(3).
				Background(m.styles.LightBG).
				Foreground(m.styles.MutedText).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(m.styles.MutedText).
				Padding(1).
				Align(lipgloss.Center)
		}
		tabs = append(tabs, tabStyle.Render(tabName))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...) + "\n"
}

func (m *ActivityDetail) setContent() {
	m.viewport.Width = m.styles.Dimensions.Width
	m.viewport.Height = m.styles.Dimensions.Height
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

	content.WriteString(m.styles.TwoColumnLayout(leftContent, rightContent, m.styles.Dimensions.Width/2))

	return content.String()
}

func (m *ActivityDetail) renderStatsCards() string {
	cardWidth := (m.styles.Dimensions.Width - 16) / 4

	cards := []string{
		m.styles.StatCard("Duration", m.activity.FormattedDuration(), m.styles.PrimaryBlue, cardWidth),
		m.styles.StatCard("Distance", m.activity.FormattedDistance(), m.styles.PrimaryGreen, cardWidth),
		m.styles.StatCard("Avg Pace", m.activity.FormattedPace(), m.styles.PrimaryOrange, cardWidth),
		m.styles.StatCard("Calories", fmt.Sprintf("%d", m.activity.Calories), m.styles.PrimaryPink, cardWidth),
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}

func (m *ActivityDetail) renderBasicMetrics() string {
	var content strings.Builder

	content.WriteString(lipgloss.NewStyle().
		Foreground(m.styles.PrimaryPurple).
		Bold(true).
		MarginBottom(1).
		Render("Activity Details"))
	content.WriteString("\n\n")

	metrics := []struct {
		label string
		value string
		color lipgloss.Color
	}{
		{"Date", m.activity.Date.Format("Monday, January 2, 2006"), m.styles.LightText},
		{"Type", strings.Title(m.activity.Type), m.styles.PrimaryBlue},
		{"Duration", m.activity.FormattedDuration(), m.styles.PrimaryGreen},
		{"Distance", m.activity.FormattedDistance(), m.styles.PrimaryOrange},
		{"Calories", fmt.Sprintf("%d kcal", m.activity.Calories), m.styles.PrimaryPink},
	}

	for _, metric := range metrics {
		content.WriteString(lipgloss.NewStyle().
			Foreground(m.styles.MutedText).
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
		Foreground(m.styles.PrimaryYellow).
		Bold(true).
		MarginBottom(1).
		Render("Performance Metrics"))
	content.WriteString("\n\n")

	metrics := []struct {
		label string
		value string
		color lipgloss.Color
	}{
		{"Avg Heart Rate", fmt.Sprintf("%d bpm", m.activity.Metrics.AvgHeartRate), m.styles.PrimaryPink},
		{"Max Heart Rate", fmt.Sprintf("%d bpm", m.activity.Metrics.MaxHeartRate), m.styles.PrimaryPink},
		{"Avg Speed", fmt.Sprintf("%.1f km/h", m.activity.Metrics.AvgSpeed), m.styles.PrimaryBlue},
		{"Elevation Gain", fmt.Sprintf("%.0f m", m.activity.Metrics.ElevationGain), m.styles.PrimaryGreen},
		{"Training Stress", fmt.Sprintf("%.1f TSS", m.activity.Metrics.TrainingStressScore), m.styles.PrimaryOrange},
		{"Recovery Time", fmt.Sprintf("%d hours", m.activity.Metrics.RecoveryTime), m.styles.PrimaryPurple},
		{"Intensity Factor", fmt.Sprintf("%.2f", m.activity.Metrics.IntensityFactor), m.styles.PrimaryYellow},
	}

	for _, metric := range metrics {
		content.WriteString(lipgloss.NewStyle().
			Foreground(m.styles.MutedText).
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
	chartsAvailable := false

	content.WriteString(lipgloss.NewStyle().
		Foreground(m.styles.PrimaryBlue).
		Bold(true).
		MarginBottom(2).
		Render("Performance Charts"))
	content.WriteString("\n\n")

	// Calculate chart dimensions based on terminal size
	chartWidth := m.viewport.Width - 12
	chartHeight := (m.viewport.Height - 15) / 3 // Divide by 3 for three charts

	// Update chart dimensions only if they've changed
	if m.hrChart.Width != chartWidth || m.hrChart.Height != chartHeight {
		m.hrChart.Width = chartWidth
		m.hrChart.Height = chartHeight
		m.powerChart.Width = chartWidth
		m.powerChart.Height = chartHeight
		m.elevationChart.Width = chartWidth
		m.elevationChart.Height = chartHeight
	}

	// Render HR chart if data exists
	if len(m.activity.Metrics.HeartRateData) > 0 {
		content.WriteString(m.hrChart.View())
		content.WriteString("\n")
		chartsAvailable = true
	}

	// Render Power chart if data exists
	if len(m.activity.Metrics.PowerData) > 0 {
		content.WriteString(m.powerChart.View())
		content.WriteString("\n")
		chartsAvailable = true
	}

	// Render Elevation chart if data exists
	if len(m.activity.Metrics.ElevationData) > 0 {
		content.WriteString(m.elevationChart.View())
		content.WriteString("\n")
		chartsAvailable = true
	}

	if !chartsAvailable {
		content.WriteString(lipgloss.NewStyle().
			Foreground(m.styles.MutedText).
			Align(lipgloss.Center).
			Width(m.viewport.Width - 8).
			Render("No chart data available for this activity"))
	} else {
		// Chart legend/info
		content.WriteString(lipgloss.NewStyle().
			Foreground(m.styles.MutedText).
			Italic(true).
			MarginTop(1).
			Render("Charts show real-time data throughout the activity duration"))
	}

	return content.String()
}

func (m *ActivityDetail) generateAnalysisCmd(forceRefresh bool) tea.Cmd {
	return func() tea.Msg {
		// Create storage and analysis clients
		analysisCache := storage.NewAnalysisCache(m.config.StoragePath)
		analysisClient := analysis.NewOpenRouterClient(m.config)

		// Check cache unless forcing refresh
		if !forceRefresh {
			cachedContent, _, err := analysisCache.GetAnalysis(m.activity.ID)
			if err == nil && cachedContent != "" {
				return AnalysisCompleteMsg{
					Analysis: cachedContent,
				}
			}
		}

		// Generate new analysis
		analysisContent, err := analysisClient.AnalyzeActivity(context.Background(), analysis.PromptParams{
			Activity: m.activity,
		})
		if err != nil {
			return AnalysisFailedMsg{
				Error: fmt.Errorf("analysis generation failed: %w", err),
			}
		}

		// Cache the analysis
		meta := storage.AnalysisMetadata{
			ActivityID:  m.activity.ID,
			GeneratedAt: time.Now(),
			ModelUsed:   m.config.OpenRouter.Model,
		}
		if err := analysisCache.StoreAnalysis(m.activity, analysisContent, meta); err != nil {
			m.logger.Warnf("Failed to cache analysis: %v", err)
		}

		return AnalysisCompleteMsg{
			Analysis: analysisContent,
		}
	}
}

func (m *ActivityDetail) renderAnalysisTab() string {
	var content strings.Builder

	content.WriteString(lipgloss.NewStyle().
		Foreground(m.styles.PrimaryGreen).
		Bold(true).
		MarginBottom(2).
		Render("AI Analysis"))
	content.WriteString("\n\n")

	if m.generating {
		spinnerView := lipgloss.JoinHorizontal(lipgloss.Left,
			m.analysisSpinner.View(),
			" "+m.analysisProgress,
		)
		content.WriteString(
			lipgloss.JoinVertical(lipgloss.Left,
				spinnerView,
				"\n\n",
				lipgloss.NewStyle().
					Foreground(m.styles.MutedText).
					Italic(true).
					Render("Analyzing workout data with AI..."),
			),
		)
	} else if m.analysis == "" {
		if m.lastError != nil {
			content.WriteString(
				lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FF0000")).
					Render(m.analysisProgress),
			)
			content.WriteString("\n\n")
			content.WriteString(
				lipgloss.NewStyle().
					Foreground(m.styles.LightText).
					Render(m.lastError.Error()),
			)
			content.WriteString("\n\n")
			content.WriteString(
				lipgloss.NewStyle().
					Foreground(m.styles.LightText).
					Render("Press 'r' to retry or 'a' to try again"),
			)
		} else {
			content.WriteString(
				lipgloss.JoinVertical(lipgloss.Left,
					lipgloss.NewStyle().
						Foreground(m.styles.MutedText).
						Align(lipgloss.Center).
						Width(m.viewport.Width-8).
						Render("No AI analysis available for this activity"),
					"\n\n",
					lipgloss.NewStyle().
						Foreground(m.styles.LightText).
						Render("Press 'a' to generate analysis"),
				),
			)
		}
	} else {
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
					Foreground(m.styles.LightText).
					Render(section))
				content.WriteString("\n\n")
				continue
			}

			title := strings.TrimSpace(parts[0])
			body := strings.TrimSpace(parts[1])

			// Use different colors for different sections
			colors := []lipgloss.Color{
				m.styles.PrimaryBlue,
				m.styles.PrimaryGreen,
				m.styles.PrimaryOrange,
				m.styles.PrimaryPink,
				m.styles.PrimaryPurple,
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
						Foreground(m.styles.LightText).
						Render("  • " + strings.TrimPrefix(line, "- ")))
				} else if strings.TrimSpace(line) != "" {
					content.WriteString(lipgloss.NewStyle().
						Foreground(m.styles.LightText).
						Render(line))
				}
				content.WriteString("\n")
			}
			content.WriteString("\n")
		}
	}

	return content.String()
}
