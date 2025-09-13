package screens

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/sstent/fitness-tui/internal/garmin"
	"github.com/sstent/fitness-tui/internal/storage"
	"github.com/sstent/fitness-tui/internal/tui/models"
)

var (
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Padding(1, 2)
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
)

type ActivityList struct {
	list         list.Model
	storage      *storage.ActivityStorage
	garminClient garmin.GarminClient
	width        int
	height       int
	statusMsg    string
	isLoading    bool
}

type activityItem struct {
	activity *models.Activity
}

func (i activityItem) FilterValue() string { return i.activity.Name }

func (i activityItem) Title() string {
	return fmt.Sprintf("%s - %s",
		i.activity.Date.Format("2006-01-02"),
		i.activity.Name)
}

func (i activityItem) Description() string {
	return fmt.Sprintf("%s  %s  %s",
		i.activity.FormattedDistance(),
		i.activity.FormattedDuration(),
		i.activity.FormattedPace())
}

func NewActivityList(storage *storage.ActivityStorage, client garmin.GarminClient) *ActivityList {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("170")).
		BorderLeftForeground(lipgloss.Color("170"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("243"))

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Activities"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().
		MarginLeft(2).
		Foreground(lipgloss.Color("62")).
		Bold(true)

	return &ActivityList{
		list:         l,
		storage:      storage,
		garminClient: client,
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
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "s":
			if !m.isLoading {
				return m, tea.Batch(m.syncActivities, m.setLoading(true))
			}
		case "enter":
			if selectedItem := m.list.SelectedItem(); selectedItem != nil {
				// TODO: Navigate to activity detail
				return m, nil
			}
		}

	case activitiesLoadedMsg:
		items := make([]list.Item, len(msg.activities))
		for i, activity := range msg.activities {
			items[i] = activityItem{activity: activity}
		}
		m.list.SetItems(items)
		return m, nil

	case loadingMsg:
		m.isLoading = bool(msg)
		return m, nil

	case syncCompleteMsg:
		m.statusMsg = fmt.Sprintf("Synced %d activities", msg.count)
		return m, tea.Batch(m.loadActivities, m.setLoading(false))

	case syncErrorMsg:
		m.statusMsg = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Red color for errors
			MarginTop(1).
			Render(fmt.Sprintf("⚠️  Sync failed: %v\nPress 's' to retry", msg.error))
		return m, m.setLoading(false)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *ActivityList) View() string {
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1).
		Render("↑↓ navigate • enter view • s sync • q back")

	status := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Render(m.statusMsg)

	if m.isLoading {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Render("Syncing with Garmin...")
	}

	return fmt.Sprintf("%s\n%s\n%s", m.list.View(), instructions, status)
}

// Messages and commands
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

	activities, err := m.garminClient.GetActivities(context.Background(), 50, &garmin.NoopLogger{})
	if err != nil {
		return syncErrorMsg{err}
	}

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
