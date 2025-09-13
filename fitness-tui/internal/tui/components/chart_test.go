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
		assert.Equal(t, "Single\n▄▄▄▄▄", view)
	})

	t.Run("multiple data points", func(t *testing.T) {
		data := []float64{10, 20, 30, 40, 50}
		chart := NewChart(data, 5, 4, "Series")
		view := chart.View()
		assert.Equal(t, "Series\n▁▂▄▆█", view)
	})

	t.Run("downsampling", func(t *testing.T) {
		data := make([]float64, 100)
		for i := range data {
			data[i] = float64(i)
		}
		chart := NewChart(data, 20, 4, "Downsample")
		view := chart.View()
		assert.Len(t, view, 20+6) // Title + chart characters
	})
}
