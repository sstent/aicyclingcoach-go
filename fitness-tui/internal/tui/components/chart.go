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

	var chart strings.Builder
	for _, value := range sampled {
		var level int
		if max == min {
			// All values are the same, use middle level
			level = 4
		} else {
			normalized := (value - min) / (max - min)
			level = int(normalized * 8)
			if level > 8 {
				level = 8
			}
			if level < 0 {
				level = 0
			}
		}
		chart.WriteString(blockChars[level])
	}

	return c.style.Render(fmt.Sprintf("%s\n%s", c.Title, chart.String()))
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
