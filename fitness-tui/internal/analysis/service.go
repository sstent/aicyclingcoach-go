package analysis

import (
	"context"
	"fmt"
	"time"

	"github.com/sstent/fitness-tui/internal/storage"
	"github.com/sstent/fitness-tui/internal/tui/models"
)

type ServiceResponse struct {
	Analysis string
	Error    error
	Duration time.Duration
}

type AnalysisService struct {
	client  *OpenRouterClient
	storage *storage.AnalysisCache
}

func NewAnalysisService(client *OpenRouterClient, storage *storage.AnalysisCache) *AnalysisService {
	return &AnalysisService{
		client:  client,
		storage: storage,
	}
}

// GetAnalysis retrieves cached analysis or generates new analysis if needed
func (s *AnalysisService) GetAnalysis(ctx context.Context, activity *models.Activity, workoutGoal string) (string, error) {
	// Check cache first
	if content, _, err := s.storage.GetAnalysis(activity.ID); err == nil {
		return content, nil
	}

	// Build prompt
	trainingContext := map[string]interface{}{
		"FTP":              activity.Metrics.FTP,
		"WorkoutType":      workoutGoal,
		"TargetZones":      activity.Metrics.TargetZones,
		"TrainingLoad":     activity.Metrics.TrainingLoad,
		"ElevationProfile": activity.Metrics.ElevationProfile,
		"FatigueLevel":     activity.Metrics.FatigueLevel,
	}

	promptParams := PromptParams{
		Activity:        activity,
		TrainingContext: trainingContext,
		Config:          s.client.config,
	}

	// Generate analysis
	analysis, err := s.client.AnalyzeActivity(ctx, promptParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate analysis: %w", err)
	}

	// Cache the result with metadata
	meta := storage.AnalysisMetadata{
		ActivityID:  activity.ID,
		GeneratedAt: time.Now(),
		ModelUsed:   s.client.config.OpenRouter.Model,
	}
	if err := s.storage.StoreAnalysis(activity, analysis, meta); err != nil {
		return "", fmt.Errorf("failed to cache analysis: %w", err)
	}

	return analysis, nil
}

// GenerateAnalysisAsync starts analysis in a goroutine and returns a channel for the result
func (s *AnalysisService) GenerateAnalysisAsync(ctx context.Context, activity *models.Activity, workoutGoal string) <-chan ServiceResponse {
	resultChan := make(chan ServiceResponse, 1)

	go func() {
		start := time.Now()
		analysis, err := s.GetAnalysis(ctx, activity, workoutGoal)
		duration := time.Since(start)

		resultChan <- ServiceResponse{
			Analysis: analysis,
			Error:    err,
			Duration: duration,
		}
	}()

	return resultChan
}
