package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sstent/fitness-tui/internal/tui/models"
)

const (
	analysisDir = "analysis"
	metaSuffix  = "-meta.json"
)

type AnalysisCache struct {
	storagePath string
}

func NewAnalysisCache(storagePath string) *AnalysisCache {
	return &AnalysisCache{
		storagePath: filepath.Join(storagePath, analysisDir),
	}
}

type AnalysisMetadata struct {
	ActivityID  string    `json:"activity_id"`
	GeneratedAt time.Time `json:"generated_at"`
	ModelUsed   string    `json:"model_used"`
	Hash        string    `json:"hash"`
}

func (c *AnalysisCache) StoreAnalysis(activity *models.Activity, content string, meta AnalysisMetadata) error {
	basePath := filepath.Join(c.storagePath, activity.ID)
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return fmt.Errorf("failed to create analysis dir: %w", err)
	}

	// Write analysis content
	contentPath := filepath.Join(basePath, "analysis.md")
	if err := os.WriteFile(contentPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write analysis: %w", err)
	}

	// Write metadata
	metaPath := filepath.Join(basePath, metaSuffix)
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metaPath, metaJSON, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

func (c *AnalysisCache) GetAnalysis(activityID string) (string, *AnalysisMetadata, error) {
	basePath := filepath.Join(c.storagePath, activityID)
	contentPath := filepath.Join(basePath, "analysis.md")
	metaPath := filepath.Join(basePath, metaSuffix)

	content, err := os.ReadFile(contentPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read analysis: %w", err)
	}

	metaJSON, err := os.ReadFile(metaPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var meta AnalysisMetadata
	if err := json.Unmarshal(metaJSON, &meta); err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return string(content), &meta, nil
}

func (c *AnalysisCache) HasFreshAnalysis(activityID string, ttl time.Duration) bool {
	basePath := filepath.Join(c.storagePath, activityID)
	metaPath := filepath.Join(basePath, metaSuffix)

	info, err := os.Stat(metaPath)
	if err != nil {
		return false
	}

	return time.Since(info.ModTime()) < ttl
}
