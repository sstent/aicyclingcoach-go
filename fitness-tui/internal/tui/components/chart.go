// internal/tui/components/chart.go
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	BlockEmpty = " "
	Block1     = "▁"
	Block2     = "▂"
	Block3     = "▃"
	Block4     = "▄"
	Block5     = "▅"
	Block6     = "▆"
	Block7     = "▇"
	Block8     = "█"
)

var blockChars = []string{BlockEmpty, Block1, Block2, Block3, Block4, Block5, Block6, Block7, Block8}

type Chart struct {
	Width    int
	Height   int
	Data     []float64
	Title    string
	Color    lipgloss.Color
	ShowAxes bool
	ShowGrid bool
	style    lipgloss.Style
}

func NewChart(data []float64, width, height int, title string) *Chart {
	return &Chart{
		Data:     data,
		Width:    width,
		Height:   height,
		Title:    title,
		Color:    lipgloss.Color("#00D4FF"), // Default bright cyan
		ShowAxes: true,
		ShowGrid: false,
		style:    lipgloss.NewStyle().Padding(0, 1),
	}
}

func NewColoredChart(data []float64, width, height int, title string, color lipgloss.Color) *Chart {
	return &Chart{
		Data:     data,
		Width:    width,
		Height:   height,
		Title:    title,
		Color:    color,
		ShowAxes: true,
		ShowGrid: false,
		style:    lipgloss.NewStyle().Padding(0, 1),
	}
}

func (c *Chart) WithColor(color lipgloss.Color) *Chart {
	c.Color = color
	return c
}

func (c *Chart) WithGrid(showGrid bool) *Chart {
	c.ShowGrid = showGrid
	return c
}

func (c *Chart) WithAxes(showAxes bool) *Chart {
	c.ShowAxes = showAxes
	return c
}

func (c *Chart) View() string {
	if len(c.Data) == 0 {
		return c.style.Render(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true).
			Render("No data available"))
	}

	min, max := findMinMax(c.Data)
	sampled := sampleData(c.Data, c.Width-8) // Account for Y-axis labels

	var content strings.Builder

	// Add title with color
	if c.Title != "" {
		titleStyle := lipgloss.NewStyle().
			Foreground(c.Color).
			Bold(true).
			MarginBottom(1)
		content.WriteString(titleStyle.Render(c.Title))
		content.WriteString("\n")
	}

	// Create the chart
	chartContent := c.renderChart(sampled, min, max)
	content.WriteString(chartContent)

	return c.style.Render(content.String())
}

func (c *Chart) renderChart(data []float64, min, max float64) string {
	if c.Height < 3 {
		// Fallback for very small charts
		return c.renderMiniChart(data, min, max)
	}

	var content strings.Builder
	chartHeight := c.Height - 1 // Reserve space for X-axis
	chartWidth := len(data)

	// Create chart grid
	rows := make([][]rune, chartHeight)
	for i := range rows {
		rows[i] = make([]rune, chartWidth)
		for j := range rows[i] {
			if c.ShowGrid && (i%2 == 0 || j%5 == 0) {
				rows[i][j] = '·'
			} else {
				rows[i][j] = ' '
			}
		}
	}

	// Plot data points
	for i, value := range data {
		if i >= chartWidth {
			break
		}

		var level int
		if max == min {
			level = chartHeight / 2
		} else {
			normalized := (value - min) / (max - min)
			level = int(normalized * float64(chartHeight-1))
		}

		if level >= chartHeight {
			level = chartHeight - 1
		}
		if level < 0 {
			level = 0
		}

		// Draw from bottom up
		rowIndex := chartHeight - 1 - level

		// Use different block characters based on the data density
		var blockIndex int
		if max == min {
			blockIndex = 3 // Use Block4 (▄) for constant data
		} else {
			normalized := (value - min) / (max - min)
			blockIndex = int(normalized * 7)
			if blockIndex > 7 {
				blockIndex = 7
			}
			if blockIndex < 0 {
				blockIndex = 0
			}
		}
		rows[rowIndex][i] = []rune(blockChars[blockIndex+1])[0]
	}

	// Render chart with Y-axis labels
	if c.ShowAxes {
		for i := 0; i < chartHeight; i++ {
			// Y-axis label
			yValue := max - (float64(i)/float64(chartHeight-1))*(max-min)
			label := fmt.Sprintf("%6.1f", yValue)

			// Y-axis line and data
			line := fmt.Sprintf("%s │", label)

			// Color the chart data
			chartLine := string(rows[i])
			coloredLine := lipgloss.NewStyle().
				Foreground(c.Color).
				Render(chartLine)

			content.WriteString(line + coloredLine + "\n")
		}

		// X-axis
		content.WriteString("       └")
		content.WriteString(strings.Repeat("─", chartWidth))
		content.WriteString("\n")

		// X-axis labels (show first, middle, last)
		spacing := chartWidth/2 - 1
		if spacing < 0 {
			spacing = 0
		}
		xAxisLabels := fmt.Sprintf("       %s%s%s",
			"0",
			strings.Repeat(" ", spacing),
			fmt.Sprintf("%d", len(c.Data)))

		if chartWidth > 10 {
			midPoint := chartWidth / 2
			leftSpacing := midPoint - 2
			rightSpacing := midPoint - 2
			if leftSpacing < 0 {
				leftSpacing = 0
			}
			if rightSpacing < 0 {
				rightSpacing = 0
			}
			xAxisLabels = fmt.Sprintf("       0%s%d%s%d",
				strings.Repeat(" ", leftSpacing),
				len(c.Data)/2,
				strings.Repeat(" ", rightSpacing),
				len(c.Data))
		}

		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Render(xAxisLabels))
	} else {
		// Simple chart without axes
		for i := 0; i < chartHeight; i++ {
			chartLine := string(rows[i])
			coloredLine := lipgloss.NewStyle().
				Foreground(c.Color).
				Render(chartLine)
			content.WriteString(coloredLine + "\n")
		}
	}

	return content.String()
}

func (c *Chart) renderMiniChart(data []float64, min, max float64) string {
	var result strings.Builder

	for _, value := range data {
		var blockIndex int
		if max == min {
			blockIndex = 4 // Middle block
		} else {
			normalized := (value - min) / (max - min)
			blockIndex = int(normalized * 7)
			if blockIndex > 7 {
				blockIndex = 7
			}
			if blockIndex < 0 {
				blockIndex = 0
			}
		}

		block := lipgloss.NewStyle().
			Foreground(c.Color).
			Render(blockChars[blockIndex+1])
		result.WriteString(block)
	}

	return result.String()
}

// Helper functions for min/max integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Helper functions remain the same
func findMinMax(data []float64) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}

	min, max := data[0], data[0]
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

func sampleData(data []float64, targetLength int) []float64 {
	if len(data) <= targetLength || targetLength <= 0 {
		return data
	}

	sampled := make([]float64, targetLength)
	ratio := float64(len(data)) / float64(targetLength)

	for i := 0; i < targetLength; i++ {
		index := int(float64(i) * ratio)
		if index >= len(data) {
			index = len(data) - 1
		}
		sampled[i] = data[index]
	}
	return sampled
}

// Sparkline creates a simple inline chart
func (c *Chart) Sparkline() string {
	if len(c.Data) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Render("─────")
	}

	minVal, maxVal := findMinMax(c.Data)
	var result strings.Builder

	for _, value := range c.Data {
		var blockIndex int
		if maxVal == minVal {
			blockIndex = 4
		} else {
			normalized := (value - minVal) / (maxVal - minVal)
			blockIndex = int(normalized * 7)
			if blockIndex > 7 {
				blockIndex = 7
			}
		}

		block := lipgloss.NewStyle().
			Foreground(c.Color).
			Render(blockChars[blockIndex+1])
		result.WriteString(block)
	}

	return result.String()
}
