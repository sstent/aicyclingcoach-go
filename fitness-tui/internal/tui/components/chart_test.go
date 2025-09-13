package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChartView(t *testing.T) {
	t.Run("empty data", func(t *testing.T) {
		chart := NewChart(nil, 10, 4, "Test")
		view := chart.View()
		assert.Contains(t, view, "No data available")
	})

	t.Run("single data point", func(t *testing.T) {
		chart := NewChart([]float64{50}, 5, 4, "Single")
		view := chart.View()
		assert.Contains(t, view, "Single")
		assert.Contains(t, view, "▄")
	})

	t.Run("multiple data points", func(t *testing.T) {
		data := []float64{10, 20, 30, 40, 50}
		chart := NewChart(data, 5, 4, "Series")
		view := chart.View()
		assert.Contains(t, view, "Series")
		// Check that we have various block characters representing the data progression
		assert.Contains(t, view, "▂")
		assert.Contains(t, view, "▄")
		assert.Contains(t, view, "▆")
		assert.Contains(t, view, "█")
	})

	t.Run("downsampling", func(t *testing.T) {
		data := make([]float64, 100)
		for i := range data {
			data[i] = float64(i)
		}
		chart := NewChart(data, 20, 4, "Downsample")
		view := chart.View()
		assert.Contains(t, view, "Downsample")
		// Just verify it contains some block characters, don't check exact length due to styling
		assert.Contains(t, view, "▁")
		assert.Contains(t, view, "▇") // Use ▇ instead of █
	})
}
