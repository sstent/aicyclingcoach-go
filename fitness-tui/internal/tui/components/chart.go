package components

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/sstent/fitness-tui/internal/types"
)

// Chart represents an ASCII chart component
type Chart struct {
	Data        []float64
	Title       string
	Width       int
	Height      int
	Color       lipgloss.Color
	Min, Max    float64
	Downsampled []types.DownsampledPoint
	Unit        string
	Mode        string // "bar" or "sparkline"
	XLabels     []string
	YMax        float64
}

// NewChart creates a new chart instance
func NewChart(data []float64, title, unit string, width, height int, color lipgloss.Color) *Chart {
	c := &Chart{
		Data:   data,
		Title:  title,
		Width:  width,
		Height: height,
		Color:  color,
		Unit:   unit,
		Mode:   "sparkline",
	}

	if len(data) > 0 {
		// Use downsampled data for min/max to improve performance
		// Using empty timestamps array since we only need value downsampling
		downsampled := types.DownsampleLTTB(data, make([]time.Time, len(data)), width)
		c.Downsampled = downsampled
		values := make([]float64, len(downsampled))
		for i, point := range downsampled {
			values[i] = point.Value
		}
		c.Min, c.Max = minMax(values)
	}
	return c
}

// NewBarChart creates a bar chart with axis labels
func NewBarChart(data []float64, title string, width, height int, color lipgloss.Color, xLabels []string, yMax float64) *Chart {
	c := NewChart(data, title, "", width, height, color)
	c.Mode = "bar"
	c.XLabels = xLabels
	c.YMax = yMax
	return c
}

// View renders the chart
func (c *Chart) View() string {
	if len(c.Data) == 0 {
		return c.renderNoData()
	}

	if c.Width <= 10 || c.Height <= 4 {
		return c.renderTooSmall()
	}

	if c.Mode == "bar" && c.YMax == 0 {
		return c.renderNoData()
	}

	// Recalculate if dimensions changed
	if len(c.Downsampled) != c.Width {
		c.Downsampled = types.DownsampleLTTB(c.Data, make([]time.Time, len(c.Data)), c.Width)
		values := make([]float64, len(c.Downsampled))
		for i, point := range c.Downsampled {
			values[i] = point.Value
		}
		c.Min, c.Max = minMax(values)
	}

	return lipgloss.NewStyle().
		MaxWidth(c.Width).
		Render(c.renderTitle() + "\n" + c.renderChart())
}

func (c *Chart) renderTitle() string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(c.Color).
		Render(c.Title)
}

func (c *Chart) renderChart() string {
	if c.Max == c.Min && c.Mode != "bar" {
		return c.renderConstantData()
	}

	if c.Mode == "bar" {
		return c.renderBarChart()
	}
	return c.renderSparkline()
}

func (c *Chart) renderBarChart() string {
	var sb strings.Builder
	chartHeight := c.Height - 3 // Reserve space for title and labels

	// Calculate scaling factor
	maxValue := c.YMax
	if maxValue == 0 {
		maxValue = c.Max
	}
	scale := float64(chartHeight-1) / maxValue

	// Y-axis labels
	yIncrement := maxValue / float64(chartHeight-1)
	for i := chartHeight - 1; i >= 0; i-- {
		label := fmt.Sprintf("%3.0f │", maxValue-(yIncrement*float64(i)))
		sb.WriteString(label)
		if i == 0 {
			sb.WriteString(" " + strings.Repeat("─", c.Width))
			break
		}

		// Draw bars
		for _, point := range c.Downsampled {
			barHeight := int(math.Round(point.Value * scale))
			if barHeight >= i {
				sb.WriteString("#")
			} else {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}

	// X-axis labels
	sb.WriteString("\n    ")
	for i, label := range c.XLabels {
		if i >= len(c.Downsampled) {
			break
		}
		if i == 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%-*s", c.Width/len(c.XLabels), label))
	}

	return sb.String()
}

func (c *Chart) renderSparkline() string {
	var sb strings.Builder
	chartHeight := c.Height - 3 // Reserve rows for title and labels

	minLabel := fmt.Sprintf("%.0f%s", c.Min, c.Unit)
	maxLabel := fmt.Sprintf("%.0f%s", c.Max, c.Unit)
	sb.WriteString(fmt.Sprintf("%5s ", maxLabel))

	for i, point := range c.Downsampled {
		if i >= c.Width {
			break
		}

		normalized := (point.Value - c.Min) / (c.Max - c.Min)
		barHeight := int(math.Round(normalized * float64(chartHeight-1)))
		sb.WriteString(c.renderBar(barHeight, chartHeight))
	}
	sb.WriteString("\n")

	// Add X axis with min label
	sb.WriteString(fmt.Sprintf("%5s ", minLabel))
	sb.WriteString(strings.Repeat("─", c.Width))
	return sb.String()
}

func (c *Chart) renderBar(height, maxHeight int) string {
	if height <= 0 {
		return " "
	}

	// Use Unicode block characters for better resolution
	blocks := []string{" ", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	index := int(float64(height) / float64(maxHeight) * 8)
	if index >= len(blocks) {
		index = len(blocks) - 1
	}

	return lipgloss.NewStyle().
		Foreground(c.Color).
		Render(blocks[index])
}

func (c *Chart) renderNoData() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(fmt.Sprintf("%s: No data", c.Title))
}

func (c *Chart) renderTooSmall() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Render(fmt.Sprintf("%s: Terminal too small", c.Title))
}

func (c *Chart) renderConstantData() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Render(fmt.Sprintf("%s: Constant value %.2f", c.Title, c.Data[0]))
}

func minMax(data []float64) (min, max float64) {
	if len(data) == 0 {
		return 0, 0
	}
	min, max = data[0], data[0]
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
