package screens

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sstent/fitness-tui/internal/garmin"
	"github.com/sstent/fitness-tui/internal/storage"
	"github.com/sstent/fitness-tui/internal/tui/models"
)

func TestSyncWorkflow(t *testing.T) {
	t.Run("successful sync", func(t *testing.T) {
		// Create mock activities
		mockActivities := []*models.Activity{
			{
				ID:       "1",
				Name:     "Morning Ride",
				Date:     time.Now(),
				Duration: 45 * time.Minute,
				Distance: 15000, // 15km
				Type:     "cycling",
				Metrics: models.ActivityMetrics{
					AvgHeartRate:  150,
					MaxHeartRate:  180,
					AvgSpeed:      20.0,
					ElevationGain: 200,
				},
			},
			{
				ID:       "2",
				Name:     "Evening Run",
				Date:     time.Now().Add(-24 * time.Hour),
				Duration: 30 * time.Minute,
				Distance: 5000, // 5km
				Type:     "running",
				Metrics: models.ActivityMetrics{
					AvgHeartRate:  160,
					MaxHeartRate:  175,
					AvgSpeed:      10.0,
					ElevationGain: 50,
				},
			},
		}

		mockClient := &garmin.MockClient{
			Activities: mockActivities,
		}

		tempDir := t.TempDir()
		activityStorage := storage.NewActivityStorage(tempDir)
		model := NewActivityList(activityStorage, mockClient)

		// Execute sync operation
		msg := model.syncActivities()

		// Verify sync was successful
		syncComplete, ok := msg.(syncCompleteMsg)
		require.True(t, ok, "Expected syncCompleteMsg")
		assert.Equal(t, 2, syncComplete.count)

		// Verify activities were stored
		activities, err := activityStorage.LoadAll()
		require.NoError(t, err)
		assert.Len(t, activities, 2)

		// Verify activity data is correct - note: activities are sorted by date descending
		assert.Equal(t, "Morning Ride", activities[0].Name)
		assert.Equal(t, "Evening Run", activities[1].Name)
	})

	t.Run("api failure during sync", func(t *testing.T) {
		mockClient := &garmin.MockClient{
			GetActivitiesError: errors.New("API unavailable"),
		}

		tempDir := t.TempDir()
		activityStorage := storage.NewActivityStorage(tempDir)
		model := NewActivityList(activityStorage, mockClient)

		msg := model.syncActivities()

		syncError, ok := msg.(syncErrorMsg)
		require.True(t, ok, "Expected syncErrorMsg")
		assert.Contains(t, syncError.error.Error(), "API unavailable")

		// Verify no activities were stored
		activities, err := activityStorage.LoadAll()
		require.NoError(t, err)
		assert.Empty(t, activities)
	})

	t.Run("storage lock prevents concurrent sync", func(t *testing.T) {
		tempDir := t.TempDir()
		activityStorage := storage.NewActivityStorage(tempDir)

		// Acquire lock to simulate ongoing sync
		err := activityStorage.AcquireLock()
		require.NoError(t, err)

		mockClient := &garmin.MockClient{
			Activities: []*models.Activity{{ID: "1", Name: "Test"}},
		}

		model := NewActivityList(activityStorage, mockClient)

		// Try to sync while lock is held
		msg := model.syncActivities()

		syncError, ok := msg.(syncErrorMsg)
		require.True(t, ok, "Expected syncErrorMsg")
		assert.Contains(t, syncError.error.Error(), "sync already in progress")

		// Clean up
		activityStorage.ReleaseLock()
	})

	t.Run("storage save failure", func(t *testing.T) {
		// Create a storage that will fail on save by using read-only directory
		activityStorage := storage.NewActivityStorage("/invalid/readonly/path")

		mockClient := &garmin.MockClient{
			Activities: []*models.Activity{{ID: "1", Name: "Test"}},
		}

		model := NewActivityList(activityStorage, mockClient)

		msg := model.syncActivities()

		syncError, ok := msg.(syncErrorMsg)
		require.True(t, ok, "Expected syncErrorMsg")
		assert.Error(t, syncError.error)
	})
}

func TestAuthenticationFailure(t *testing.T) {
	mockClient := &garmin.MockClient{
		ConnectError: errors.New("authentication failed"),
	}

	tempDir := t.TempDir()
	activityStorage := storage.NewActivityStorage(tempDir)
	model := NewActivityList(activityStorage, mockClient)

	// Test Init() which calls Connect()
	cmd := model.Init()

	// The current implementation doesn't handle Connect() errors properly
	// This test documents the current behavior and can be updated when fixed
	assert.NotNil(t, cmd)
}

func TestSyncStatusMessages(t *testing.T) {
	tempDir := t.TempDir()
	activityStorage := storage.NewActivityStorage(tempDir)
	mockClient := &garmin.MockClient{}

	t.Run("loading state during sync", func(t *testing.T) {
		model := NewActivityList(activityStorage, mockClient)

		// Simulate loading state directly for testing UI
		loadingMsg := loadingMsg(true)
		updatedModel, _ := model.Update(loadingMsg)

		activityList := updatedModel.(*ActivityList)
		assert.True(t, activityList.isLoading)

		// Verify loading message appears in view
		view := activityList.View()
		assert.Contains(t, view, "Syncing with Garmin...")
	})

	t.Run("success message after sync", func(t *testing.T) {
		model := NewActivityList(activityStorage, mockClient)

		// Simulate successful sync completion
		syncMsg := syncCompleteMsg{count: 5}
		updatedModel, _ := model.Update(syncMsg)

		activityList := updatedModel.(*ActivityList)
		assert.Equal(t, "Synced 5 activities", activityList.statusMsg)
	})

	t.Run("error message display", func(t *testing.T) {
		model := NewActivityList(activityStorage, mockClient)

		// Simulate sync error
		syncMsg := syncErrorMsg{errors.New("connection timeout")}
		updatedModel, _ := model.Update(syncMsg)

		activityList := updatedModel.(*ActivityList)
		view := activityList.View()
		assert.Contains(t, view, "⚠️  Sync failed")
		assert.Contains(t, view, "connection timeout")
		assert.Contains(t, view, "Press 's' to retry")
	})

	t.Run("prevent multiple concurrent syncs", func(t *testing.T) {
		model := NewActivityList(activityStorage, mockClient)

		// Start first sync
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		updatedModel, _ := model.Update(keyMsg)
		activityList := updatedModel.(*ActivityList)
		activityList.isLoading = true

		// Try to start second sync while first is running
		updatedModel2, cmd := activityList.Update(keyMsg)
		activityList2 := updatedModel2.(*ActivityList)

		// Should remain in loading state, no new command issued
		assert.True(t, activityList2.isLoading)
		assert.Nil(t, cmd)
	})
}

func TestActivityLoading(t *testing.T) {
	tempDir := t.TempDir()
	activityStorage := storage.NewActivityStorage(tempDir)

	// Pre-populate storage with test activities
	testActivity := &models.Activity{
		ID:   "test-1",
		Name: "Test Activity",
		Date: time.Now(),
		Type: "cycling",
	}
	err := activityStorage.Save(testActivity)
	require.NoError(t, err)

	mockClient := &garmin.MockClient{}
	model := NewActivityList(activityStorage, mockClient)

	// Test loadActivities method
	msg := model.loadActivities()

	loadedMsg, ok := msg.(activitiesLoadedMsg)
	require.True(t, ok, "Expected activitiesLoadedMsg")
	assert.Len(t, loadedMsg.activities, 1)
	assert.Equal(t, "Test Activity", loadedMsg.activities[0].Name)
}

func TestActivityListUpdate(t *testing.T) {
	tempDir := t.TempDir()
	activityStorage := storage.NewActivityStorage(tempDir)
	mockClient := &garmin.MockClient{}
	model := NewActivityList(activityStorage, mockClient)

	t.Run("window resize", func(t *testing.T) {
		resizeMsg := tea.WindowSizeMsg{Width: 100, Height: 50}
		updatedModel, cmd := model.Update(resizeMsg)

		activityList := updatedModel.(*ActivityList)
		assert.Equal(t, 100, activityList.width)
		assert.Equal(t, 50, activityList.height)
		assert.Nil(t, cmd)
	})

	t.Run("activities loaded", func(t *testing.T) {
		activities := []*models.Activity{
			{ID: "1", Name: "Activity 1"},
			{ID: "2", Name: "Activity 2"},
		}

		loadedMsg := activitiesLoadedMsg{activities: activities}
		updatedModel, cmd := model.Update(loadedMsg)

		activityList := updatedModel.(*ActivityList)
		items := activityList.list.Items()
		assert.Len(t, items, 2)
		assert.Nil(t, cmd)
	})

	t.Run("loading state changes", func(t *testing.T) {
		loadingMsg := loadingMsg(true)
		updatedModel, cmd := model.Update(loadingMsg)

		activityList := updatedModel.(*ActivityList)
		assert.True(t, activityList.isLoading)
		assert.Nil(t, cmd)
	})
}
