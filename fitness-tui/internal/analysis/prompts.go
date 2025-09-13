package analysis

import (
	"fmt"

	"github.com/sstent/fitness-tui/internal/tui/models"
)

const analysisPromptTemplate = `Analyze this cycling activity for training effectiveness:

**Activity Details**
- Name: %s
- Date: %s
- Duration: %s
- Distance: %s
- Elevation Gain: %.0fm
- Average Heart Rate: %d bpm
- Max Heart Rate: %d bpm
- Average Speed: %.1f km/h

**Training Focus**
The athlete intended this workout to be: %s

**Analysis Request**
Provide structured feedback in markdown format covering:
1. Training goal achievement assessment
2. Heart rate zone analysis
3. Pacing strategy evaluation
4. Recovery recommendations
5. Suggested improvements for similar future workouts

Keep responses concise and focused on actionable insights. Use bullet points and headings for organization.`

func BuildAnalysisPrompt(activity *models.Activity, workoutGoal string) string {
	return fmt.Sprintf(analysisPromptTemplate,
		activity.Name,
		activity.Date.Format("2006-01-02 15:04"),
		activity.FormattedDuration(),
		activity.FormattedDistance(),
		activity.Metrics.ElevationGain,
		activity.Metrics.AvgHeartRate,
		activity.Metrics.MaxHeartRate,
		activity.Metrics.AvgSpeed,
		workoutGoal,
	)
}
