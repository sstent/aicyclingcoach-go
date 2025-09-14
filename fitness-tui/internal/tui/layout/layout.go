// internal/tui/layout/layout.go
package layout

import (
	"github.com/charmbracelet/lipgloss"
)

// Color palette with bright, distinctive colors
var (
	// Primary colors
	PrimaryBlue   = lipgloss.Color("#00D4FF") // Bright cyan
	PrimaryCyan   = lipgloss.Color("#00FFFF") // Pure cyan
	PrimaryPurple = lipgloss.Color("#B300FF") // Bright purple
	PrimaryGreen  = lipgloss.Color("#00FF88") // Bright green
	PrimaryOrange = lipgloss.Color("#FF8800") // Bright orange
	PrimaryPink   = lipgloss.Color("#FF0088") // Bright pink
	PrimaryYellow = lipgloss.Color("#FFD700") // Gold

	// Background colors
	DarkBG  = lipgloss.Color("#1a1a1a")
	LightBG = lipgloss.Color("#2a2a2a")
	CardBG  = lipgloss.Color("#333333")

	// Text colors
	WhiteText = lipgloss.Color("#FFFFFF")
	LightText = lipgloss.Color("#CCCCCC")
	MutedText = lipgloss.Color("#888888")
)

// Layout represents the main layout structure
type Layout struct {
	Width  int
	Height int
}

// NewLayout creates a new layout with given dimensions
func NewLayout(width, height int) *Layout {
	return &Layout{
		Width:  width,
		Height: height,
	}
}

// MainContainer creates the primary container style
func (l *Layout) MainContainer() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(l.Width).
		Height(l.Height).
		Background(DarkBG).
		Padding(1)
}

// HeaderPanel creates a header panel with title and navigation
func (l *Layout) HeaderPanel(title string, breadcrumb string) string {
	headerStyle := lipgloss.NewStyle().
		Width(l.Width-4).
		Height(3).
		Background(PrimaryBlue).
		Foreground(WhiteText).
		Bold(true).
		Padding(1, 2).
		MarginBottom(1)

	titleStyle := lipgloss.NewStyle().
		Foreground(WhiteText).
		Bold(true)

	breadcrumbStyle := lipgloss.NewStyle().
		Foreground(LightText).
		Italic(true)

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(title),
		breadcrumbStyle.Render(breadcrumb),
	)

	return headerStyle.Render(content)
}

// SidePanel creates a colorful side panel
func (l *Layout) SidePanel(title string, content string, color lipgloss.Color, width int) string {
	panelStyle := lipgloss.NewStyle().
		Width(width).
		Height(l.Height-8).
		Background(CardBG).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(1, 2).
		MarginRight(1)

	titleStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		MarginBottom(1)

	contentStyle := lipgloss.NewStyle().
		Foreground(LightText).
		Width(width - 6)

	panelContent := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(title),
		contentStyle.Render(content),
	)

	return panelStyle.Render(panelContent)
}

// MainPanel creates the main content panel
func (l *Layout) MainPanel(content string, remainingWidth int) string {
	panelStyle := lipgloss.NewStyle().
		Width(remainingWidth).
		Height(l.Height-8).
		Background(CardBG).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryPurple).
		Padding(1, 2)

	contentStyle := lipgloss.NewStyle().
		Foreground(LightText).
		Width(remainingWidth - 6)

	return panelStyle.Render(contentStyle.Render(content))
}

// StatCard creates a statistics card with bright colors
func (l *Layout) StatCard(title string, value string, color lipgloss.Color, width int) string {
	cardStyle := lipgloss.NewStyle().
		Width(width).
		Height(4).
		Background(CardBG).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(1).
		MarginRight(1).
		MarginBottom(1)

	titleStyle := lipgloss.NewStyle().
		Foreground(MutedText).
		Bold(false).
		Align(lipgloss.Center)

	valueStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Align(lipgloss.Center)

	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render(title),
		valueStyle.Render(value),
	)

	return cardStyle.Render(content)
}

// NavigationBar creates a bottom navigation bar
func (l *Layout) NavigationBar(items []NavItem, activeIndex int) string {
	navStyle := lipgloss.NewStyle().
		Width(l.Width-4).
		Height(2).
		Background(LightBG).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryGreen).
		Padding(0, 2).
		MarginTop(1)

	var navItems []string
	for i, item := range items {
		if i == activeIndex {
			activeStyle := lipgloss.NewStyle().
				Foreground(PrimaryGreen).
				Bold(true).
				Background(DarkBG).
				Padding(0, 2)
			navItems = append(navItems, activeStyle.Render(item.Label))
		} else {
			inactiveStyle := lipgloss.NewStyle().
				Foreground(MutedText).
				Padding(0, 2)
			navItems = append(navItems, inactiveStyle.Render(item.Label))
		}
	}

	content := lipgloss.JoinHorizontal(lipgloss.Center, navItems...)
	return navStyle.Render(content)
}

// ActivityCard creates a card for activity items
func (l *Layout) ActivityCard(name, date, duration, distance string, isSelected bool) string {
	width := (l.Width - 12) / 2 // Two columns

	var cardStyle lipgloss.Style
	if isSelected {
		cardStyle = lipgloss.NewStyle().
			Width(width).
			Height(6).
			Background(CardBG).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryOrange).
			Padding(1, 2).
			MarginRight(1).
			MarginBottom(1)
	} else {
		cardStyle = lipgloss.NewStyle().
			Width(width).
			Height(6).
			Background(CardBG).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(MutedText).
			Padding(1, 2).
			MarginRight(1).
			MarginBottom(1)
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(WhiteText).
		Bold(true).
		MarginBottom(1)

	metaStyle := lipgloss.NewStyle().
		Foreground(LightText)

	dateStyle := lipgloss.NewStyle().
		Foreground(PrimaryBlue)

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(name),
		dateStyle.Render(date),
		metaStyle.Render(duration+" • "+distance),
	)

	return cardStyle.Render(content)
}

// ChartPanel creates a panel for charts
func (l *Layout) ChartPanel(title string, chartContent string, color lipgloss.Color) string {
	panelStyle := lipgloss.NewStyle().
		Width(l.Width-8).
		Background(CardBG).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(1, 2).
		MarginBottom(1)

	titleStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		MarginBottom(1)

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(title),
		chartContent,
	)

	return panelStyle.Render(content)
}

// HelpText creates styled help text
func (l *Layout) HelpText(text string) string {
	return lipgloss.NewStyle().
		Foreground(MutedText).
		Italic(true).
		Align(lipgloss.Center).
		Width(l.Width - 4).
		Render(text)
}

// NavItem represents a navigation item
type NavItem struct {
	Label string
	Key   string
}

// TwoColumnLayout creates a two-column layout
func (l *Layout) TwoColumnLayout(leftContent, rightContent string, leftWidth int) string {
	rightWidth := l.Width - leftWidth - 6

	leftPanel := lipgloss.NewStyle().
		Width(leftWidth).
		Height(l.Height-8).
		Background(CardBG).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryPink).
		Padding(1, 2).
		MarginRight(2)

	rightPanel := lipgloss.NewStyle().
		Width(rightWidth).
		Height(l.Height-8).
		Background(CardBG).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryYellow).
		Padding(1, 2)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		leftPanel.Render(leftContent),
		rightPanel.Render(rightContent),
	)
}
