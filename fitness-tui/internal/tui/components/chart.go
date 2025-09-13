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
	Width  int
	Height int
	Data   []float64
	Title  string
	style  lipgloss.Style
}

func NewChart(data []float64, width, height int, title string) *Chart {
	return &Chart{
		Data:   data,
		Width:  width,
		Height: height,
		Title:  title,
		style:  lipgloss.NewStyle().Padding(0, 1),
	}
}

func (c *Chart) View() string {
	if len(c.Data) == 0 {
		return c.style.Render("No data available")
	}

	min, max := findMinMax(c.Data)
	sampled := sampleData(c.Data, c.Width)

	// Create chart rows from top to bottom
	rows := make([]string, c.Height)
	for i := range rows {
		rows[i] = strings.Repeat(" ", c.Width)
	}

	// Fill chart based on values
	for i, value := range sampled {
		if max == min {
			// Handle case where all values are equal
			level := c.Height / 2
			if level >= c.Height {
				level = c.Height - 1
			}
			rows[level] = replaceAtIndex(rows[level], blockChars[7], i)
		} else {
			normalized := (value - min) / (max - min)
			level := int(normalized * float64(c.Height-1))
			if level >= c.Height {
				level = c.Height - 1
			}
			if level < 0 {
				level = 0
			}
			// Draw from bottom up
			rowIndex := c.Height - 1 - level
			rows[rowIndex] = replaceAtIndex(rows[rowIndex], blockChars[7], i)
		}
	}

	// Add Y-axis labels
	chartWithLabels := ""
	if c.Height > 3 {
		chartWithLabels += fmt.Sprintf("%5.1f ┤\n", max)
		for i := 1; i < c.Height-1; i++ {
			chartWithLabels += "     │ " + rows[i] + "\n"
		}
		chartWithLabels += fmt.Sprintf("%5.1f ┤ %s", min, rows[c.Height-1])
	} else {
		// Fallback for small heights
		for _, row := range rows {
			chartWithLabels += row + "\n"
		}
	}

	// Add X-axis title
	return c.style.Render(fmt.Sprintf("%s\n%s", c.Title, chartWithLabels))
}

// Helper function to replace character at index
func replaceAtIndex(in string, r string, i int) string {
	out := []rune(in)
	out[i] = []rune(r)[0]
	return string(out)
}

func findMinMax(data []float64) (float64, float64) {
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
	if len(data) <= targetLength {
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
