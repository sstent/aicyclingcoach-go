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
	ad := &ActivityDetail{
		activity:       activity,
		analysis:       analysis,
		viewport:       vp,
		styles:         styles,
		hrChart:        components.NewChart(activity.Metrics.HeartRateData, 40, 4, "Heart Rate (bpm)"),
		elevationChart: components.NewChart(activity.Metrics.ElevationData, 40, 4, "Elevation (m)"),
	}
	ad.setContent()
	return ad
}

func (m *ActivityDetail) Init() tea.Cmd {
	// Initialize content immediately instead of waiting for WindowSizeMsg
	m.setContent()
	return nil
}

func (m *ActivityDetail) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport = viewport.New(msg.Width, msg.Height-4)
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
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1).
		Render("esc back • q quit")

	return fmt.Sprintf("%s\n%s", m.viewport.View(), instructions)
}

func (m *ActivityDetail) setContent() {
	var content strings.Builder

	// Debug: Check if activity is nil
	if m.activity == nil {
		content.WriteString("Activity data is nil!")
		m.viewport.SetContent(m.styles.Viewport.Render(content.String()))
		fmt.Println("DEBUG: ActivityDetail.setContent() - activity is nil")
		return
	}

	fmt.Printf("DEBUG: ActivityDetail.setContent() - Rendering activity: %s\n", m.activity.Name)
	fmt.Printf("DEBUG: ActivityDetail.setContent() - Duration: %v, Distance: %.2f\n", m.activity.Duration, m.activity.Distance)
	fmt.Printf("DEBUG: ActivityDetail.setContent() - Metrics: AvgHR=%d, MaxHR=%d, AvgSpeed=%.2f\n", m.activity.Metrics.AvgHeartRate, m.activity.Metrics.MaxHeartRate, m.activity.Metrics.AvgSpeed)

	// Debug info at top
	content.WriteString(fmt.Sprintf("DEBUG: Viewport W=%d H=%d, Activity: %s\n", m.width, m.height, m.activity.Name))
	content.WriteString("\n")

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
		charts := lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.hrChart.View(),
			lipgloss.NewStyle().Width(2).Render(" "),
			m.elevationChart.View(),
		)
		content.WriteString(charts)
		content.WriteString("\n\n")
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

	m.viewport.SetContent(m.styles.Viewport.Render(content.String()))
}
