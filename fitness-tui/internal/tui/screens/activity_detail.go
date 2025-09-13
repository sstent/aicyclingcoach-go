package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/sstent/fitness-tui/internal/garmin"
	"github.com/sstent/fitness-tui/internal/tui/components"
	"github.com/sstent/fitness-tui/internal/tui/models"
)

type BackToListMsg struct{}

type ActivityDetail struct {
	activity       *models.Activity
	analysis       string
	viewport       viewport.Model
	width          int
	height         int
	styles         *Styles
	hrChart        *components.Chart
	elevationChart *components.Chart
	logger         garmin.Logger
	ready          bool // Tracks if viewport has been initialized
}

type Styles struct {
	Title     lipgloss.Style
	Subtitle  lipgloss.Style
	StatName  lipgloss.Style
	StatValue lipgloss.Style
	Analysis  lipgloss.Style
	Viewport  lipgloss.Style
}

func NewActivityDetail(activity *models.Activity, analysis string, logger garmin.Logger) *ActivityDetail {
	if logger == nil {
		logger = &garmin.NoopLogger{}
	}

	styles := &Styles{
		Title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")),
		Subtitle:  lipgloss.NewStyle().Foreground(lipgloss.Color("8")).MarginTop(1),
		StatName:  lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		StatValue: lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
		Analysis:  lipgloss.NewStyle().MarginTop(1),
		Viewport:  lipgloss.NewStyle().Padding(0, 1),
	}

	vp := viewport.New(80, 20) // Default dimensions
	ad := &ActivityDetail{
		activity:       activity,
		analysis:       analysis,
		viewport:       vp,
		styles:         styles,
		logger:         logger,
		hrChart:        components.NewChart(activity.Metrics.HeartRateData, 40, 4, "Heart Rate (bpm)"),
		elevationChart: components.NewChart(activity.Metrics.ElevationData, 40, 4, "Elevation (m)"),
	}
	ad.setContent()
	return ad
}

func (m *ActivityDetail) Init() tea.Cmd {
	// Request window size to get proper dimensions
	return tea.Batch(
		func() tea.Msg { return tea.WindowSizeMsg{Width: 80, Height: 24} },
	)
}

func (m *ActivityDetail) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport = viewport.New(msg.Width, msg.Height-2)
		m.ready = true
		m.setContent()
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "b", "q": // Add 'q' key for quitting/going back
			return m, func() tea.Msg { return BackToListMsg{} }
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

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("esc back • q quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		m.viewport.View(),
		instructions,
	)
}

func (m *ActivityDetail) setContent() {
	var content strings.Builder

	// Debug: Check if activity is nil
	if m.activity == nil {
		content.WriteString("Activity data is nil!")
		m.viewport.SetContent(m.styles.Viewport.Render(content.String()))
		m.logger.Debugf("ActivityDetail.setContent() - activity is nil")
		return
	}

	m.logger.Debugf("ActivityDetail.setContent() - Rendering activity: %s", m.activity.Name)
	m.logger.Debugf("ActivityDetail.setContent() - Duration: %v, Distance: %.2f", m.activity.Duration, m.activity.Distance)
	m.logger.Debugf("ActivityDetail.setContent() - Metrics: AvgHR=%d, MaxHR=%d, AvgSpeed=%.2f", m.activity.Metrics.AvgHeartRate, m.activity.Metrics.MaxHeartRate, m.activity.Metrics.AvgSpeed)

	// Debug info at top

	// Activity Details
	content.WriteString(m.styles.Title.Render(m.activity.Name))
	content.WriteString("\n\n")

	// Activity Details with two-column layout
	content.WriteString(m.styles.Subtitle.Render("Activity Details"))
	content.WriteString("\n")

	// First row
	content.WriteString(fmt.Sprintf("%s %s   %s %s\n",
		m.styles.StatName.Render("Date:"),
		m.styles.StatValue.Render(m.activity.Date.Format("2006-01-02")),
		m.styles.StatName.Render("Type:"),
		m.styles.StatValue.Render(m.activity.Type),
	))

	// Second row
	content.WriteString(fmt.Sprintf("%s %s   %s %s\n",
		m.styles.StatName.Render("Duration:"),
		m.styles.StatValue.Render(m.activity.FormattedDuration()),
		m.styles.StatName.Render("Distance:"),
		m.styles.StatValue.Render(m.activity.FormattedDistance()),
	))

	// Third row
	content.WriteString(fmt.Sprintf("%s %s   %s %s\n",
		m.styles.StatName.Render("Calories:"),
		m.styles.StatValue.Render(fmt.Sprintf("%d kcal", m.activity.Calories)),
		m.styles.StatName.Render("Elevation Gain:"),
		m.styles.StatValue.Render(fmt.Sprintf("%.0f m", m.activity.Metrics.ElevationGain)),
	))

	// Fourth row
	content.WriteString(fmt.Sprintf("%s %s   %s %s\n",
		m.styles.StatName.Render("Avg HR:"),
		m.styles.StatValue.Render(fmt.Sprintf("%d bpm", m.activity.Metrics.AvgHeartRate)),
		m.styles.StatName.Render("Max HR:"),
		m.styles.StatValue.Render(fmt.Sprintf("%d bpm", m.activity.Metrics.MaxHeartRate)),
	))

	// Fifth row (Training Load metrics)
	content.WriteString(fmt.Sprintf("%s %s   %s %s\n",
		m.styles.StatName.Render("Training Stress:"),
		m.styles.StatValue.Render(fmt.Sprintf("%.1f", m.activity.Metrics.TrainingStress)),
		m.styles.StatName.Render("Recovery Time:"),
		m.styles.StatValue.Render(fmt.Sprintf("%d hours", m.activity.Metrics.RecoveryTime)),
	))

	// Sixth row (Intensity Factor and Speed)
	content.WriteString(fmt.Sprintf("%s %s   %s %s\n",
		m.styles.StatName.Render("Intensity Factor:"),
		m.styles.StatValue.Render(fmt.Sprintf("%.1f", m.activity.Metrics.IntensityFactor)),
		m.styles.StatName.Render("Avg Speed:"),
		m.styles.StatValue.Render(fmt.Sprintf("%.1f km/h", m.activity.Metrics.AvgSpeed)),
	))

	// Seventh row (Pace)
	content.WriteString(fmt.Sprintf("%s %s\n",
		m.styles.StatName.Render("Avg Pace:"),
		m.styles.StatValue.Render(fmt.Sprintf("%s/km", m.activity.FormattedPace())),
	))

	// Charts Section
	content.WriteString(m.styles.Subtitle.Render("Performance Charts"))
	content.WriteString("\n")

	// Only show charts if we have data
	if len(m.activity.Metrics.HeartRateData) > 0 || len(m.activity.Metrics.ElevationData) > 0 {
		// Calculate available height for charts (about 1/3 of screen height)
		chartHeight := max(6, (m.height-20)/3)
		chartWidth := max(30, m.width-4)

		m.hrChart.Width = chartWidth
		m.hrChart.Height = chartHeight
		m.elevationChart.Width = chartWidth
		m.elevationChart.Height = chartHeight

		// Render charts with spacing
		if len(m.activity.Metrics.HeartRateData) > 0 {
			content.WriteString("\n" + m.hrChart.View() + "\n")
		}
		if len(m.activity.Metrics.ElevationData) > 0 {
			content.WriteString("\n" + m.elevationChart.View() + "\n")
		}
	}

	// Analysis Section with formatted output
	if m.analysis != "" {
		content.WriteString(m.styles.Analysis.Render(
			m.styles.Subtitle.Render("AI Analysis"),
		))
		content.WriteString("\n")

		// Split analysis into sections
		sections := strings.Split(m.analysis, "## ")
		for _, section := range sections {
			if strings.TrimSpace(section) == "" {
				continue
			}

			// Split section into title and content
			parts := strings.SplitN(section, "\n", 2)
			if len(parts) < 2 {
				content.WriteString(m.styles.StatValue.Render(section))
				continue
			}

			title := strings.TrimSpace(parts[0])
			body := strings.TrimSpace(parts[1])

			// Render section title
			content.WriteString(m.styles.Title.Render(title))
			content.WriteString("\n")

			// Format bullet points
			lines := strings.Split(body, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "- ") {
					content.WriteString("• ")
					content.WriteString(strings.TrimPrefix(line, "- "))
				} else {
					content.WriteString(line)
				}
				content.WriteString("\n")
			}
			content.WriteString("\n")
		}
	}

	m.viewport.SetContent(content.String())
}

// Helper function for max since it's not available in older Go versions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
