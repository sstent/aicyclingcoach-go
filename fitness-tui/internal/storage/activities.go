package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sstent/fitness-tui/internal/tui/models"
)

type ActivityStorage struct {
	dataDir  string
	lockPath string
}

func NewActivityStorage(dataDir string) *ActivityStorage {
	activitiesDir := filepath.Join(dataDir, "activities")
	os.MkdirAll(activitiesDir, 0755)

	return &ActivityStorage{
		dataDir:  dataDir,
		lockPath: filepath.Join(dataDir, "sync.lock"),
	}
}

// AcquireLock tries to create an exclusive lock file
func (s *ActivityStorage) AcquireLock() error {
	file, err := os.OpenFile(s.lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("sync already in progress")
		}
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	file.Close()
	return nil
}

// ReleaseLock removes the lock file
func (s *ActivityStorage) ReleaseLock() error {
	if err := os.Remove(s.lockPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	return nil
}

func (s *ActivityStorage) Save(activity *models.Activity) error {
	filename := fmt.Sprintf("%s-%s.json",
		activity.Date.Format("2006-01-02"),
		sanitizeFilename(activity.Name))
	targetPath := filepath.Join(s.dataDir, "activities", filename)

	data, err := json.MarshalIndent(activity, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal activity: %w", err)
	}

	// Atomic write using temp file and rename
	tmpFile, err := os.CreateTemp(filepath.Dir(targetPath), "tmp-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write activity data: %w", err)
	}

	// Sync to ensure write completes before rename
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), targetPath); err != nil {
		return fmt.Errorf("failed to atomically replace activity file: %w", err)
	}

	return nil
}

func (s *ActivityStorage) LoadAll() ([]*models.Activity, error) {
	activitiesDir := filepath.Join(s.dataDir, "activities")

	files, err := os.ReadDir(activitiesDir)
	if err != nil {
		return nil, err
	}

	var activities []*models.Activity
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		activity, err := s.loadActivity(filepath.Join(activitiesDir, file.Name()))
		if err != nil {
			continue // Skip invalid files
		}
		activities = append(activities, activity)
	}

	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Date.After(activities[j].Date)
	})

	return activities, nil
}

func (s *ActivityStorage) loadActivity(path string) (*models.Activity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var activity models.Activity
	if err := json.Unmarshal(data, &activity); err != nil {
		return nil, err
	}

	return &activity, nil
}

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "-", "\\", "-", ":", "-", "*", "-",
		"?", "-", "\"", "-", "<", "-", ">", "-",
		"|", "-", " ", "-",
	)
	return replacer.Replace(name)
}
