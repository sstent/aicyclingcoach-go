// internal/tui/screens/activity_list.go
package screens

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/sstent/fitness-tui/internal/garmin"
	"github.com/sstent/fitness-tui/internal/storage"
	"github.com/sstent/fitness-tui/internal/tui/layout"
	"github.com/sstent/fitness-tui/internal/tui/models"
)

type ActivityList struct {
	activities            []*models.Activity
	totalGarminActivities int // Added for sync status
	storage               *storage.ActivityStorage
	garminClient          garmin.GarminClient
	layout                *layout.Layout
	selectedIndex         int
	statusMsg             string
	isLoading             bool
	currentPage           int
	scrollOffset          int
}

func NewActivityList(storage *storage.ActivityStorage, client garmin.GarminClient) *ActivityList {
	return &ActivityList{
		storage:      storage,
		garminClient: client,
		layout:       layout.NewLayout(80, 24), // Default size
	}
}

func (m *ActivityList) Init() tea.Cmd {
	// Initialize Garmin connection synchronously for now
	m.garminClient.Connect(&garmin.NoopLogger{})
	return m.loadActivities
}

func (m *ActivityList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.layout = layout.NewLayout(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "s":
			if !m.isLoading {
				return m, tea.Batch(m.syncActivities, m.setLoading(true))
			}
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
				// Calculate visible rows based on current layout
				availableHeight := m.layout.Height - 3 - 2 - 2 - 1 - 4
				visibleRows := availableHeight - 1
				m.updateScrollOffset(visibleRows)
			}
		case "down", "j":
			if m.selectedIndex < len(m.activities)-1 {
				m.selectedIndex++
				// Calculate visible rows based on current layout
				availableHeight := m.layout.Height - 3 - 2 - 2 - 1 - 4
				visibleRows := availableHeight - 1
				m.updateScrollOffset(visibleRows)
			}
		case "pgup":
			// Calculate visible rows based on current layout
			availableHeight := m.layout.Height - 3 - 2 - 2 - 1 - 4
			visibleRows := availableHeight - 1
			m.selectedIndex = max(0, m.selectedIndex-visibleRows)
			m.updateScrollOffset(visibleRows)
		case "pgdown":
			// Calculate visible rows based on current layout
			availableHeight := m.layout.Height - 3 - 2 - 2 - 1 - 4
			visibleRows := availableHeight - 1
			m.selectedIndex = min(len(m.activities)-1, m.selectedIndex+visibleRows)
			m.updateScrollOffset(visibleRows)
		case "enter":
			if len(m.activities) > 0 && m.selectedIndex < len(m.activities) {
				activity := m.activities[m.selectedIndex]
				return m, func() tea.Msg {
					return ActivitySelectedMsg{Activity: activity}
				}
			}
		}

	case activitiesLoadedMsg:
		m.activities = msg.activities
		if m.selectedIndex >= len(m.activities) {
			m.selectedIndex = max(0, len(m.activities)-1)
		}
		// Calculate visible rows based on current layout
		availableHeight := m.layout.Height - 3 - 2 - 2 - 1 - 4
		visibleRows := availableHeight - 1
		m.updateScrollOffset(visibleRows)
		return m, nil

	case loadingMsg:
		m.isLoading = bool(msg)
		return m, nil

	case syncCompleteMsg:
		m.statusMsg = fmt.Sprintf("✓ Synced %d activities", msg.count)
		return m, tea.Batch(m.loadActivities, m.setLoading(false))

	case syncErrorMsg:
		m.statusMsg = fmt.Sprintf("⚠️ Sync failed: %v", msg.error)
		return m, m.setLoading(false)
	}

	return m, nil
}

func (m *ActivityList) updateScrollOffset(visibleRows int) {
	if m.selectedIndex < m.scrollOffset {
		m.scrollOffset = m.selectedIndex
	} else if m.selectedIndex >= m.scrollOffset+visibleRows {
		m.scrollOffset = m.selectedIndex - visibleRows + 1
	}

	// Ensure scroll offset doesn't go negative
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func (m *ActivityList) View() string {
	var content strings.Builder

	// Header (fixed height)
	headerHeight := 3
	breadcrumb := "Home > Activities"
	if m.isLoading {
		breadcrumb += " (Syncing...)"
	}
	content.WriteString(m.layout.HeaderPanel("Fitness Activities", breadcrumb))

	// Removed stats panel per user request

	// Calculate available height for main content
	navHeight := 2
	helpHeight := 1
	padding := 4 // Total vertical padding
	availableHeight := m.layout.Height - headerHeight - navHeight - helpHeight - padding

	// Calculate column widths
	summaryWidth := 40
	listWidth := m.layout.Width - summaryWidth - 2

	if listWidth < 30 {
		// Single column - full width
		listContent := lipgloss.NewStyle().
			Height(availableHeight).
			Render(m.renderActivityList())
		content.WriteString(m.layout.MainPanel(listContent, m.layout.Width-6))
	} else {
		// Two columns - use full available height
		listContent := lipgloss.NewStyle().
			Width(listWidth).
			Height(availableHeight).
			Render(m.renderActivityList())

		summaryContent := lipgloss.NewStyle().
			Width(summaryWidth).
			Height(availableHeight).
			MarginLeft(1).
			Render(m.renderSummaryPanel())

		content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listContent, summaryContent))
	}

	// Navigation bar
	navItems := []layout.NavItem{
		{Label: "Activities", Key: "a"},
		{Label: "Routes", Key: "r"},
		{Label: "Plans", Key: "p"},
		{Label: "Workouts", Key: "w"},
		{Label: "Analytics", Key: "n"},
	}

	// Add sync status to nav bar
	syncStatus := fmt.Sprintf("Synced: %d/%d", len(m.activities), m.totalGarminActivities)
	navBar := m.layout.NavigationBar(navItems, 0)
	statusStyle := lipgloss.NewStyle().Foreground(layout.MutedText).Align(lipgloss.Right)
	statusText := statusStyle.Render(syncStatus)
	navBar = lipgloss.JoinHorizontal(lipgloss.Bottom, navBar, statusText)
	content.WriteString(navBar)

	// Help text - removed left/right navigation hints
	helpText := "↑↓ navigate • enter select • s sync • q quit"
	if m.statusMsg != "" {
		helpText = m.statusMsg + " • " + helpText
	}
	content.WriteString(m.layout.HelpText(helpText))

	return m.layout.MainContainer().Render(content.String())
}

// Removed stats panel per user request

func (m *ActivityList) renderActivityList() string {
	if len(m.activities) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(layout.MutedText).
			Align(lipgloss.Center).
			Width(m.layout.Width*2/3 - 6).
			Height(10)
		return emptyStyle.Render("No activities found\nPress 's' to sync with Garmin")
	}

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().
		Foreground(layout.WhiteText).
		Bold(true).
		MarginBottom(1).
		Render("Recent Activities"))
	content.WriteString("\n")

	// Calculate dynamic visible rows based on available space
	availableHeight := m.layout.Height - 3 - 2 - 2 - 1 - 4 // header, stats, nav, help, padding
	visibleRows := availableHeight - 1                     // subtract 1 for title
	if m.scrollOffset > 0 {
		visibleRows-- // reserve space for "more above" indicator
	}
	if m.scrollOffset+visibleRows < len(m.activities) {
		visibleRows-- // reserve space for "more below" indicator
	}

	// Ensure at least 1 visible row
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Calculate visible range based on scroll offset
	startIdx := m.scrollOffset
	endIdx := min(startIdx+visibleRows, len(m.activities))

	// Activity type color mapping
	typeColors := map[string]lipgloss.Color{
		"cycling":  layout.PrimaryBlue,
		"running":  layout.PrimaryGreen,
		"swimming": layout.PrimaryCyan,
		"hiking":   layout.PrimaryOrange,
		"walking":  layout.PrimaryYellow,
	}

	for i := startIdx; i < endIdx; i++ {
		activity := m.activities[i]
		isSelected := (i == m.selectedIndex)

		// Get color for activity type, default to white
		color, ok := typeColors[strings.ToLower(activity.Type)]
		if !ok {
			color = layout.WhiteText
		}

		// Format activity line
		dateStr := activity.Date.Format("2006-01-02")
		typeStr := activity.Type
		nameStr := activity.Name

		// Apply coloring
		dateStyle := lipgloss.NewStyle().Foreground(layout.MutedText)
		typeStyle := lipgloss.NewStyle().Foreground(color).Bold(true)
		nameStyle := lipgloss.NewStyle().Foreground(layout.LightText)

		if isSelected {
			dateStyle = dateStyle.Bold(true)
			typeStyle = typeStyle.Bold(true).Underline(true)
			nameStyle = nameStyle.Bold(true)
			content.WriteString("> ")
		} else {
			content.WriteString("  ")
		}

		content.WriteString(dateStyle.Render(dateStr))
		content.WriteString(" ")
		content.WriteString(typeStyle.Render(typeStr))
		content.WriteString(" ")
		content.WriteString(nameStyle.Render(nameStr))
		content.WriteString("\n")
	}

	// Scroll indicators
	if startIdx > 0 {
		content.WriteString(lipgloss.NewStyle().
			Foreground(layout.PrimaryBlue).
			Align(lipgloss.Center).
			Render("↑ More activities above"))
		content.WriteString("\n")
	}

	if endIdx < len(m.activities) {
		content.WriteString(lipgloss.NewStyle().
			Foreground(layout.PrimaryBlue).
			Align(lipgloss.Center).
			Render("↓ More activities below"))
		content.WriteString("\n")
	}

	return content.String()
}

func (m *ActivityList) renderSummaryPanel() string {
	var content strings.Builder

	// Activity Summary
	content.WriteString(lipgloss.NewStyle().
		Foreground(layout.WhiteText).
		Bold(true).
		MarginBottom(1).
		Render("Activity Summary"))
	content.WriteString("\n\n")

	if len(m.activities) > 0 && m.selectedIndex < len(m.activities) {
		activity := m.activities[m.selectedIndex]

		// Selected activity details
		content.WriteString(lipgloss.NewStyle().
			Foreground(layout.PrimaryYellow).
			Bold(true).
			Render(activity.Name))
		content.WriteString("\n")

		content.WriteString(lipgloss.NewStyle().
			Foreground(layout.LightText).
			Render(activity.Date.Format("Monday, January 2, 2006")))
		content.WriteString("\n\n")

		// Key metrics
		metrics := []struct {
			label string
			value string
			color lipgloss.Color
		}{
			{"Duration", activity.FormattedDuration(), layout.PrimaryGreen},
			{"Distance", activity.FormattedDistance(), layout.PrimaryBlue},
			{"Avg Pace", activity.FormattedPace(), layout.PrimaryOrange},
			{"Calories", fmt.Sprintf("%d kcal", activity.Calories), layout.PrimaryPink},
			{"Avg HR", fmt.Sprintf("%d bpm", activity.Metrics.AvgHeartRate), layout.PrimaryPurple},
			{"Elevation", fmt.Sprintf("%.0f m", activity.Metrics.ElevationGain), layout.PrimaryGreen},
		}

		for _, metric := range metrics {
			content.WriteString(lipgloss.NewStyle().
				Foreground(layout.MutedText).
				Render(metric.label + ": "))
			content.WriteString(lipgloss.NewStyle().
				Foreground(metric.color).
				Bold(true).
				Render(metric.value))
			content.WriteString("\n")
		}

		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().
			Foreground(layout.MutedText).
			Italic(true).
			Render("Press Enter to view detailed analysis"))
	} else {
		content.WriteString(lipgloss.NewStyle().
			Foreground(layout.MutedText).
			Render("Select an activity to view summary"))
	}

	return content.String()
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Messages and commands (unchanged)
type ActivitySelectedMsg struct {
	Activity *models.Activity
}

type activitiesLoadedMsg struct {
	activities []*models.Activity
}

type loadingMsg bool
type syncErrorMsg struct{ error }
type syncCompleteMsg struct{ count int }

func (m *ActivityList) loadActivities() tea.Msg {
	activities, err := m.storage.LoadAll()
	if err != nil {
		return syncErrorMsg{err}
	}
	return activitiesLoadedMsg{activities: activities}
}

func (m *ActivityList) syncActivities() tea.Msg {
	if err := m.storage.AcquireLock(); err != nil {
		return syncErrorMsg{err}
	}
	defer m.storage.ReleaseLock()

	// Increase limit to 10,000 activities
	activities, err := m.garminClient.GetActivities(context.Background(), 10000, &garmin.NoopLogger{})
	if err != nil {
		return syncErrorMsg{err}
	}

	// Update total count for status display
	m.totalGarminActivities = len(activities)

	for _, activity := range activities {
		if err := m.storage.Save(activity); err != nil {
			return syncErrorMsg{err}
		}
	}
	return syncCompleteMsg{count: len(activities)}
}

func (m *ActivityList) setLoading(isLoading bool) tea.Cmd {
	return func() tea.Msg {
		return loadingMsg(isLoading)
	}
}
