package screens

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/sstent/fitness-tui/internal/config"
	"github.com/sstent/fitness-tui/internal/garmin"
	"github.com/sstent/fitness-tui/internal/tui/models"
	"github.com/sstent/fitness-tui/internal/tui/styles"
	"github.com/stretchr/testify/assert"
)

func TestActivityDetail_NilActivityHandling(t *testing.T) {
	// Setup with nil activity
	ad := NewActivityDetail(nil, "", &config.Config{}, &garmin.NoopLogger{})
	ad.setContent()

	// Verify nil check in setContent
	assert.Contains(t, ad.viewport.View(), "Activity data is nil!", "Should show nil activity message")

	// Verify nil check in renderOverviewTab
	content := ad.renderOverviewTab()
	assert.Contains(t, content, "Activity data is nil!", "Should show nil activity in overview")
}

func TestActivityDetail_NilMetricsHandling(t *testing.T) {
	// Create activity with nil metrics
	activity := &models.Activity{
		Name: "Test Activity",
		Metrics: &models.ActivityMetrics{
			HeartRateData: []float64{},
			PowerData:     []float64{},
			ElevationData: []float64{},
		},
	}

	ad := NewActivityDetail(activity, "", &config.Config{}, &garmin.NoopLogger{})

	// Switch to charts tab
	ad.currentTab = 1
	ad.setContent()
	content := ad.viewport.View()

	assert.Contains(t, content, "No chart data available", "Should handle nil metrics")
	assert.NotContains(t, content, "Heart Rate", "Should not render HR chart")
	assert.NotContains(t, content, "Power", "Should not render power chart")
	assert.NotContains(t, content, "Elevation", "Should not render elevation chart")
}

func TestActivityDetail_AnalysisErrorStates(t *testing.T) {
	activity := &models.Activity{Name: "Test Activity"}
	ad := NewActivityDetail(activity, "", &config.Config{}, &garmin.NoopLogger{})
	ad.currentTab = 2

	// Test error display
	ad.lastError = fmt.Errorf("test error")
	ad.setContent()
	content := ad.viewport.View()
	assert.Contains(t, content, "Analysis failed", "Should show error state")
	assert.Contains(t, content, "test error", "Should show error message")
	assert.Contains(t, content, "Press 'r' to retry", "Should show retry prompt")

	// Test generating state
	ad.generating = true
	ad.analysisProgress = "Generating..."
	ad.setContent()
	content = ad.viewport.View()
	assert.Contains(t, content, "Generating...", "Should show progress message")
}

func BenchmarkActivityDetail_Render(b *testing.B) {
	activity := &models.Activity{
		Name: "Benchmark Activity",
		Metrics: &models.ActivityMetrics{
			HeartRateData: make([]float64, 1000),
			PowerData:     make([]float64, 1000),
			ElevationData: make([]float64, 1000),
		},
	}

	ad := NewActivityDetail(activity, "", &config.Config{}, &garmin.NoopLogger{})
	// Mock styles and layout for rendering
	ad.styles = &styles.Styles{
		CardBG:        "#ffffff",
		PrimaryOrange: "#e67e22",
		MutedText:     "#7f8c8d",
		LightText:     "#bdc3c7",
	}
	ad.viewport = viewport.New(100, 40)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ad.setContent()
		_ = ad.View()
	}
}
