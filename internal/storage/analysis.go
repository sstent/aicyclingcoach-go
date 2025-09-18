package storage

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sstent/fitness-tui/internal/tui/models"
	"gopkg.in/yaml.v3"
)

const (
	analysisDir = "analysis"
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
	ActivityID  string    `yaml:"activity_id"`
	GeneratedAt time.Time `yaml:"generated_at"`
	ModelUsed   string    `yaml:"model_used"`
	Hash        string    `yaml:"hash"`
}

func (c *AnalysisCache) StoreAnalysis(activity *models.Activity, content string, meta AnalysisMetadata) error {
	basePath := filepath.Join(c.storagePath, activity.ID)
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return fmt.Errorf("failed to create analysis dir: %w", err)
	}

	// Encode metadata as YAML
	metaYAML, err := yaml.Marshal(&meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create hybrid document with YAML front matter
	var hybridContent bytes.Buffer
	hybridContent.WriteString("---\n")
	hybridContent.Write(metaYAML)
	hybridContent.WriteString("---\n\n")
	hybridContent.WriteString(content)

	// Write hybrid document
	contentPath := filepath.Join(basePath, "analysis.md")
	if err := os.WriteFile(contentPath, hybridContent.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write analysis: %w", err)
	}

	return nil
}

func (c *AnalysisCache) GetAnalysis(activityID string) (string, *AnalysisMetadata, error) {
	basePath := filepath.Join(c.storagePath, activityID)
	contentPath := filepath.Join(basePath, "analysis.md")

	data, err := os.ReadFile(contentPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read analysis: %w", err)
	}

	// Split YAML front matter from content
	parts := strings.SplitN(string(data), "---", 3)
	if len(parts) < 3 {
		return "", nil, fmt.Errorf("invalid analysis format")
	}

	// Parse YAML metadata
	var meta AnalysisMetadata
	if err := yaml.Unmarshal([]byte(parts[1]), &meta); err != nil {
		return "", nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// The rest is markdown content
	content := strings.TrimSpace(parts[2])
	return content, &meta, nil
}

func (c *AnalysisCache) HasFreshAnalysis(activityID string, ttl time.Duration) bool {
	basePath := filepath.Join(c.storagePath, activityID)
	contentPath := filepath.Join(basePath, "analysis.md")

	info, err := os.Stat(contentPath)
	if err != nil {
		return false
	}

	return time.Since(info.ModTime()) < ttl
}
