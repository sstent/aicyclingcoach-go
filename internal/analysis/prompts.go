package analysis

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sstent/fitness-tui/internal/config"
	"github.com/sstent/fitness-tui/internal/tui/models"
)

type PromptParams struct {
	Activity        *models.Activity
	Goal            string
	Locale          string
	TrainingContext interface{} `json:"training_context,omitempty"`
	Config          *config.Config
}

func GeneratePrompt(params PromptParams) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("Analyze this %s workout from %s:\n",
		params.Activity.Type, params.Activity.Date.Format("2006-01-02")))
	prompt.WriteString(fmt.Sprintf("- Duration: %s\n", params.Activity.Duration))
	prompt.WriteString(fmt.Sprintf("- Distance: %.1f km\n", params.Activity.Distance/1000))
	prompt.WriteString(fmt.Sprintf("- Elevation: %.0f m\n", params.Activity.Metrics.ElevationGain))
	prompt.WriteString(fmt.Sprintf("- Avg Power: %.0fW\n", params.Activity.Metrics.AvgPower))
	prompt.WriteString(fmt.Sprintf("- Avg HR: %d bpm\n", params.Activity.Metrics.AvgHeartRate))
	prompt.WriteString("\nTraining Context:\n")
	if params.TrainingContext != nil {
		contextJSON, _ := json.Marshal(params.TrainingContext)
		prompt.WriteString(string(contextJSON))
	}
	prompt.WriteString("\n\nProvide structured analysis in this format:\n")
	prompt.WriteString("- Summary: [concise overview]\n")
	prompt.WriteString("- Strengths: [2-3 bullet points]\n")
	prompt.WriteString("- Improvements: [2-3 actionable suggestions]")

	return prompt.String()
}
