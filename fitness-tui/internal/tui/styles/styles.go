package styles

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	Dimensions struct {
		Width  int
		Height int
	}

	PrimaryBlue     lipgloss.Color
	PrimaryGreen    lipgloss.Color
	PrimaryOrange   lipgloss.Color
	PrimaryPink     lipgloss.Color
	PrimaryPurple   lipgloss.Color
	PrimaryYellow   lipgloss.Color
	LightBG         lipgloss.Color
	DarkBG          lipgloss.Color
	CardBG          lipgloss.Color
	MutedText       lipgloss.Color
	LightText       lipgloss.Color
	PrimaryText     lipgloss.Color
	HeaderPanel     lipgloss.Style
	MainPanel       lipgloss.Style
	NavigationBar   func([]NavItem, int) string
	HelpText        lipgloss.Style
	MainContainer   lipgloss.Style
	StatCard        func(string, string, lipgloss.Color, int) string
	TwoColumnLayout func(string, string, int) string
}

func NewStyles() *Styles {
	s := &Styles{}

	s.PrimaryBlue = lipgloss.Color("#3498db")
	s.PrimaryGreen = lipgloss.Color("#2ecc71")
	s.PrimaryOrange = lipgloss.Color("#e67e22")
	s.PrimaryPink = lipgloss.Color("#e84393")
	s.PrimaryPurple = lipgloss.Color("#9b59b6")
	s.PrimaryYellow = lipgloss.Color("#f1c40f")
	s.LightBG = lipgloss.Color("#ecf0f1")
	s.DarkBG = lipgloss.Color("#2c3e50")
	s.CardBG = lipgloss.Color("#ffffff")
	s.MutedText = lipgloss.Color("#7f8c8d")
	s.LightText = lipgloss.Color("#bdc3c7")
	s.PrimaryText = lipgloss.Color("#2c3e50")

	// Initialize dimensions with default values
	s.Dimensions.Width = 80
	s.Dimensions.Height = 24

	s.HeaderPanel = lipgloss.NewStyle().
		Foreground(s.PrimaryText).
		Background(s.PrimaryBlue).
		Bold(true).
		Padding(0, 1).
		Width(s.Dimensions.Width)

	s.HelpText = lipgloss.NewStyle().
		Foreground(s.MutedText).
		Padding(0, 1)

	s.MainContainer = lipgloss.NewStyle().
		Padding(1, 2)

	s.StatCard = func(title, value string, color lipgloss.Color, width int) string {
		return lipgloss.NewStyle().
			Background(s.CardBG).
			Foreground(color).
			Padding(1).
			Width(width).
			Render(fmt.Sprintf("%s\n%s", title, value))
	}

	s.TwoColumnLayout = func(left, right string, width int) string {
		return lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(width/2).Render(left),
			lipgloss.NewStyle().Width(width/2).Render(right),
		)
	}

	s.NavigationBar = func(items []NavItem, activeIdx int) string {
		var navItems []string
		for i, item := range items {
			style := lipgloss.NewStyle().
				Padding(0, 1).
				Foreground(s.MutedText)

			if i == activeIdx {
				style = style.
					Foreground(s.PrimaryText).
					Bold(true)
			}
			navItems = append(navItems, style.Render(item.Label))
		}
		return lipgloss.JoinHorizontal(lipgloss.Left, navItems...)
	}

	return s
}

// NavItem defines a navigation bar item
type NavItem struct {
	Label string
	Key   string
}
