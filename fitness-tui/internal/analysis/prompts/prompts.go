package prompts

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/sstent/fitness-tui/internal/tui/models"
)

// PromptTemplate defines a structured prompt for LLM analysis
type PromptTemplate struct {
	System    string
	User      string
	Functions []string
}

// GetAnalysisPrompt returns the prompt template for activity analysis
func GetAnalysisPrompt(activity *models.Activity) PromptTemplate {
	return PromptTemplate{
		System: "You are an experienced cycling coach analyzing a training session. Provide concise, actionable feedback focusing on training execution, intensity control, and areas for improvement. Use markdown formatting for headings and lists. Structure your response with the following sections: Workout Execution, Intensity Control, Areas for Improvement, and Recommendations.",
		User: fmt.Sprintf(`## Activity Overview
- **Name**: %s
- **Type**: %s (%s)
- **Date**: %s
- **Duration**: %s
- **Distance**: %.1f km
- **Elevation Gain**: %.0f m

## Key Metrics
- **Avg Heart Rate**: %d bpm
- **Avg Power**: %.0f w
- **TSS**: %.0f
- **Intensity Factor (IF)**: %.1f
- **Normalized Power (NP)**: %.0f w

## Training Context
- **Training Goal**: %s

## Analysis Request
Please provide a detailed analysis of this activity in relation to the athlete's training goal. Focus on:
1. Workout execution: Did the athlete achieve the intended goal?
2. Intensity control: How well did the athlete manage their effort?
3. Areas for improvement: What could be done better in future similar workouts?
4. Recommendations: Any adjustments for upcoming training?`,
			activity.Name,
			activity.Type,
			activity.ActivityType,
			activity.Date.Format("2006-01-02 15:04"),
			activity.Duration,
			activity.Distance/1000, // Convert meters to km
			activity.Metrics.ElevationGain,
			activity.Metrics.AvgHeartRate,
			activity.Metrics.AvgPower,
			activity.Metrics.TrainingStressScore,
			activity.Metrics.IntensityFactor,
			activity.Metrics.NormalizedPower,
			activity.Metrics.TargetZones,
		),
		Functions: []string{},
	}
}

// RenderTemplate renders the prompt template with activity data
func RenderTemplate(tmpl string, data interface{}) (string, error) {
	t, err := template.New("prompt").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
