package components

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sstent/aicyclingcoach-go/fitness-tui/internal/types"
)

// Chart represents an ASCII chart component
type Chart struct {
	Data        []float64
	Title       string
	Width       int
	Height      int
	Color       lipgloss.Color
	downsampler *types.Downsampler
}

// NewChart creates a new Chart instance
func NewChart(data []float64, title string) *Chart {
	return &Chart{
		Data:        data,
		Title:       title,
		Width:       0, // Will be set based on terminal size
		Height:      10,
		Color:       lipgloss.Color("39"), // Default blue
		downsampler: types.NewDownsampler(),
	}
}

// WithSize sets the chart dimensions
func (c *Chart) WithSize(width, height int) *Chart {
	c.Width = width
	c.Height = height
	return c
}

// WithColor sets the chart color
func (c *Chart) WithColor(color lipgloss.Color) *Chart {
	c.Color = color
	return c
}

// View renders the chart
func (c *Chart) View() string {
	if len(c.Data) == 0 {
		return fmt.Sprintf("%s\nNo data available", c.Title)
	}

	// Downsample data if needed
	processedData := c.downsampler.Process(c.Data, c.Width)

	// Normalize data to chart height
	min, max := minMax(processedData)
	normalized := normalize(processedData, min, max, c.Height-1)

	// Build chart
	var sb strings.Builder
	sb.WriteString(c.Title + "\n")

	// Create Y-axis labels
	yLabels := createYAxisLabels(min, max, c.Height-1)

	for i := c.Height - 1; i >= 0; i-- {
		if i == 0 {
			sb.WriteString("└") // Bottom-left corner
		} else if i == c.Height-1 {
			sb.WriteString("↑") // Top axis indicator
		} else {
			sb.WriteString("│") // Y-axis line
		}

		// Add Y-axis label
		if i < len(yLabels) {
			sb.WriteString(yLabels[i])
		} else {
			sb.WriteString(" ")
		}

		// Add chart bars
		for j := 0; j < len(normalized); j++ {
			if i == 0 {
				sb.WriteString("─") // X-axis
			} else {
				if normalized[j] >= float64(i) {
					sb.WriteString("█") // Full block
				} else {
					// Gradient blocks based on fractional part
					frac := normalized[j] - math.Floor(normalized[j])
					if normalized[j] >= float64(i-1) && frac > 0.75 {
						sb.WriteString("▇")
					} else if normalized[j] >= float64(i-1) && frac > 0.5 {
						sb.WriteString("▅")
					} else if normalized[j] >= float64(i-1) && frac > 0.25 {
						sb.WriteString("▃")
					} else if normalized[j] >= float64(i-1) && frac > 0 {
						sb.WriteString("▁")
					} else {
						sb.WriteString(" ")
					}
				}
			}
		}
		sb.WriteString("\n")
	}

	// Add X-axis title
	sb.WriteString(" " + strings.Repeat(" ", len(yLabels[0])+1) + "→ Time\n")

	// Apply color styling
	style := lipgloss.NewStyle().Foreground(c.Color)
	return style.Render(sb.String())
}

// minMax finds min and max values in a slice
func minMax(data []float64) (min, max float64) {
	if len(data) == 0 {
		return 0, 0
	}
	min = data[0]
	max = data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

// normalize scales values to fit within chart height
func normalize(data []float64, min, max float64, height int) []float64 {
	if max == min || height <= 0 {
		return make([]float64, len(data))
	}

	scale := float64(height) / (max - min)
	normalized := make([]float64, len(data))
	for i, v := range data {
		normalized[i] = (v - min) * scale
	}
	return normalized
}

// createYAxisLabels creates labels for Y-axis
func createYAxisLabels(min, max float64, height int) []string {
	labels := make([]string, height+1)
	step := (max - min) / float64(height)

	for i := 0; i <= height; i++ {
		value := min + float64(i)*step
		label := fmt.Sprintf("%.0f", value)
		// Pad to consistent width (5 characters)
		if len(label) < 5 {
			label = strings.Repeat(" ", 5-len(label)) + label
		}
		labels[height-i] = label
	}
	return labels
}
