# Fitness TUI Implementation Guide

## Table of Contents
1. [Project Setup](#project-setup)
2. [Development Environment](#development-environment)
3. [Core Dependencies](#core-dependencies)
4. [Project Structure](#project-structure)
5. [Implementation Phases](#implementation-phases)
6. [Detailed Implementation Steps](#detailed-implementation-steps)
7. [Testing Strategy](#testing-strategy)
8. [Deployment](#deployment)

## Project Setup

### Initialize Go Module
```bash
mkdir fitness-tui
cd fitness-tui
go mod init github.com/yourusername/fitness-tui
```

### Create Directory Structure
```bash
mkdir -p {cmd,internal/{tui/{screens,components,models},garmin,analysis,storage,charts,config}}
mkdir -p {test,docs,examples}
touch cmd/main.go
touch internal/tui/app.go
```

### Git Setup
```bash
git init
echo "# Fitness TUI" > README.md
```

Create `.gitignore`:
```gitignore
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
fitness-tui

# Test binary
*.test

# Output of the go coverage tool
*.out

# Go workspace file
go.work

# IDE
.vscode/
.idea/

# Config files with credentials
config.yaml
*.env

# Build artifacts
dist/
build/
```

## Development Environment

### Required Go Version
- Go 1.21+ (for better error handling and performance)

### Development Tools
```bash
# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/air-verse/air@latest  # Hot reload during development
```

### IDE Configuration
Add to `.vscode/settings.json`:
```json
{
    "go.useLanguageServer": true,
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v"],
    "go.testTimeout": "10s"
}
```

## Core Dependencies

### Add Dependencies
```bash
# TUI Framework
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest

# Garmin Integration (check latest version of go-garth)
go get github.com/sstent/go-garth@latest

# HTTP Client for OpenRouter
go get github.com/go-resty/resty/v2@latest

# Configuration
go get github.com/spf13/viper@latest
go get github.com/spf13/cobra@latest

# GPX Processing
go get github.com/tkrajina/gpxgo/gpx@latest

# File System Operations
go get github.com/fsnotify/fsnotify@latest

# Testing
go get github.com/stretchr/testify@latest
```

### Verify Dependencies
```bash
go mod tidy
go mod verify
```

## Project Structure

```
fitness-tui/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   ├── config.go          # Configuration management
│   │   └── defaults.go        # Default configuration values
│   ├── tui/
│   │   ├── app.go             # Main TUI application
│   │   ├── screens/
│   │   │   ├── activity_list.go
│   │   │   ├── activity_detail.go
│   │   │   ├── route_list.go
│   │   │   ├── plan_list.go
│   │   │   └── help.go
│   │   ├── components/
│   │   │   ├── chart.go       # ASCII chart components
│   │   │   ├── list.go        # List components
│   │   │   └── modal.go       # Modal dialogs
│   │   └── models/
│   │       ├── activity.go
│   │       ├── route.go
│   │       ├── plan.go
│   │       └── analysis.go
│   ├── garmin/
│   │   ├── client.go          # Garmin API client wrapper
│   │   ├── sync.go            # Activity synchronization
│   │   └── auth.go            # Authentication handling
│   ├── analysis/
│   │   ├── llm.go             # OpenRouter LLM client
│   │   ├── cache.go           # Analysis caching
│   │   └── prompts.go         # LLM prompt templates
│   ├── storage/
│   │   ├── files.go           # File-based storage operations
│   │   ├── activities.go      # Activity storage management
│   │   ├── routes.go          # Route storage management
│   │   └── migrations.go      # Data format migrations
│   └── charts/
│       ├── ascii.go           # ASCII chart generation
│       └── sparkline.go       # Sparkline utilities
├── test/
│   ├── fixtures/              # Test data files
│   └── integration/           # Integration tests
├── docs/
├── examples/
│   └── sample-config.yaml
└── README.md
```

## Implementation Phases

### Phase 1: Core Foundation (Week 1-2)
- [ ] Project setup and configuration management
- [ ] Basic TUI structure with bubbletea
- [ ] File storage system
- [ ] Garmin client integration

### Phase 2: Activity Management (Week 3-4)
- [ ] Activity data models
- [ ] Activity list screen
- [ ] Activity detail screen
- [ ] Basic ASCII charts

### Phase 3: LLM Analysis (Week 5-6)
- [ ] OpenRouter integration
- [ ] Analysis caching system
- [ ] Analysis display in TUI

### Phase 4: Routes & Polish (Week 7-8)
- [ ] GPX file handling
- [ ] Route management screens
- [ ] Error handling and edge cases
- [ ] Documentation and testing

## Detailed Implementation Steps

### Step 1: Configuration System

#### `internal/config/config.go`
```go
package config

import (
    "os"
    "path/filepath"
    
    "github.com/spf13/viper"
)

type Config struct {
    Garmin struct {
        Username string `mapstructure:"username"`
        Password string `mapstructure:"password"`
    } `mapstructure:"garmin"`
    
    OpenRouter struct {
        APIKey    string `mapstructure:"api_key"`
        Model     string `mapstructure:"model"`
        BaseURL   string `mapstructure:"base_url"`
    } `mapstructure:"openrouter"`
    
    Storage struct {
        DataDir string `mapstructure:"data_dir"`
    } `mapstructure:"storage"`
}

func Load() (*Config, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, err
    }
    
    configDir := filepath.Join(homeDir, ".fitness-tui")
    if err := os.MkdirAll(configDir, 0755); err != nil {
        return nil, err
    }
    
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(configDir)
    
    // Set defaults
    setDefaults()
    
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            // Create default config file
            if err := viper.SafeWriteConfig(); err != nil {
                return nil, err
            }
        } else {
            return nil, err
        }
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }
    
    return &config, nil
}

func setDefaults() {
    homeDir, _ := os.UserHomeDir()
    viper.SetDefault("storage.data_dir", filepath.Join(homeDir, ".fitness-tui"))
    viper.SetDefault("openrouter.model", "deepseek/deepseek-r1-05028:free")
    viper.SetDefault("openrouter.base_url", "https://openrouter.ai/api/v1")
}
```

### Step 2: Data Models

#### `internal/tui/models/activity.go`
```go
package models

import (
    "time"
    
    "github.com/tkrajina/gpxgo/gpx"
)

type Activity struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        string                 `json:"type"`
    Date        time.Time              `json:"date"`
    Duration    time.Duration          `json:"duration"`
    Distance    float64                `json:"distance"` // meters
    Metrics     ActivityMetrics        `json:"metrics"`
    GPXData     *gpx.GPX              `json:"-"`
    GPXPath     string                 `json:"gpx_path,omitempty"`
    Analysis    *Analysis              `json:"analysis,omitempty"`
}

type ActivityMetrics struct {
    AvgHeartRate    int     `json:"avg_heart_rate,omitempty"`
    MaxHeartRate    int     `json:"max_heart_rate,omitempty"`
    AvgPace         float64 `json:"avg_pace,omitempty"`      // seconds per km
    AvgSpeed        float64 `json:"avg_speed,omitempty"`     // km/h
    ElevationGain   float64 `json:"elevation_gain,omitempty"` // meters
    ElevationLoss   float64 `json:"elevation_loss,omitempty"` // meters
    Calories        int     `json:"calories,omitempty"`
    AvgCadence      int     `json:"avg_cadence,omitempty"`
    AvgPower        int     `json:"avg_power,omitempty"`
    Temperature     float64 `json:"temperature,omitempty"`   // celsius
}

type Analysis struct {
    ActivityID   string    `json:"activity_id"`
    WorkoutType  string    `json:"workout_type"`
    GeneratedAt  time.Time `json:"generated_at"`
    Content      string    `json:"content"`
    Insights     []string  `json:"insights"`
    FilePath     string    `json:"file_path"`
}

// Helper methods
func (a *Activity) FormattedDuration() string {
    hours := int(a.Duration.Hours())
    minutes := int(a.Duration.Minutes()) % 60
    seconds := int(a.Duration.Seconds()) % 60
    
    if hours > 0 {
        return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
    }
    return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func (a *Activity) FormattedDistance() string {
    km := a.Distance / 1000
    if km >= 10 {
        return fmt.Sprintf("%.1fkm", km)
    }
    return fmt.Sprintf("%.2fkm", km)
}

func (a *Activity) FormattedPace() string {
    if a.Metrics.AvgPace <= 0 {
        return "--:--"
    }
    
    minutes := int(a.Metrics.AvgPace / 60)
    seconds := int(a.Metrics.AvgPace) % 60
    return fmt.Sprintf("%d:%02d/km", minutes, seconds)
}
```

### Step 3: Storage System

#### `internal/storage/activities.go`
```go
package storage

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "sort"
    "time"
    
    "github.com/yourusername/fitness-tui/internal/tui/models"
)

type ActivityStorage struct {
    dataDir string
}

func NewActivityStorage(dataDir string) *ActivityStorage {
    activitiesDir := filepath.Join(dataDir, "activities")
    os.MkdirAll(activitiesDir, 0755)
    
    return &ActivityStorage{
        dataDir: dataDir,
    }
}

func (s *ActivityStorage) Save(activity *models.Activity) error {
    activitiesDir := filepath.Join(s.dataDir, "activities")
    filename := fmt.Sprintf("%s-%s.json", 
        activity.Date.Format("2006-01-02"), 
        sanitizeFilename(activity.Name))
    
    filepath := filepath.Join(activitiesDir, filename)
    
    data, err := json.MarshalIndent(activity, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal activity: %w", err)
    }
    
    if err := ioutil.WriteFile(filepath, data, 0644); err != nil {
        return fmt.Errorf("failed to write activity file: %w", err)
    }
    
    return nil
}

func (s *ActivityStorage) LoadAll() ([]*models.Activity, error) {
    activitiesDir := filepath.Join(s.dataDir, "activities")
    
    files, err := ioutil.ReadDir(activitiesDir)
    if err != nil {
        if os.IsNotExist(err) {
            return []*models.Activity{}, nil
        }
        return nil, err
    }
    
    var activities []*models.Activity
    
    for _, file := range files {
        if filepath.Ext(file.Name()) != ".json" {
            continue
        }
        
        activity, err := s.loadActivity(filepath.Join(activitiesDir, file.Name()))
        if err != nil {
            // Log error but continue loading other activities
            continue
        }
        
        activities = append(activities, activity)
    }
    
    // Sort by date (newest first)
    sort.Slice(activities, func(i, j int) bool {
        return activities[i].Date.After(activities[j].Date)
    })
    
    return activities, nil
}

func (s *ActivityStorage) loadActivity(filePath string) (*models.Activity, error) {
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return nil, err
    }
    
    var activity models.Activity
    if err := json.Unmarshal(data, &activity); err != nil {
        return nil, err
    }
    
    return &activity, nil
}

func sanitizeFilename(name string) string {
    // Replace invalid filename characters
    replacer := strings.NewReplacer(
        "/", "-",
        "\\", "-",
        ":", "-",
        "*", "-",
        "?", "-",
        "\"", "-",
        "<", "-",
        ">", "-",
        "|", "-",
        " ", "-",
    )
    return replacer.Replace(name)
}
```

### Step 4: TUI Application Structure

#### `internal/tui/app.go`
```go
package tui

import (
    "fmt"
    "log"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    
    "github.com/yourusername/fitness-tui/internal/config"
    "github.com/yourusername/fitness-tui/internal/tui/screens"
    "github.com/yourusername/fitness-tui/internal/storage"
)

type Screen int

const (
    ScreenMain Screen = iota
    ScreenActivityList
    ScreenActivityDetail
    ScreenRouteList
    ScreenPlanList
    ScreenHelp
)

type App struct {
    config          *config.Config
    activityStorage *storage.ActivityStorage
    
    currentScreen   Screen
    screens         map[Screen]tea.Model
    
    width  int
    height int
    
    statusMessage string
    errorMessage  string
}

func NewApp(cfg *config.Config) *App {
    activityStorage := storage.NewActivityStorage(cfg.Storage.DataDir)
    
    app := &App{
        config:          cfg,
        activityStorage: activityStorage,
        currentScreen:   ScreenMain,
        screens:         make(map[Screen]tea.Model),
    }
    
    // Initialize screens
    app.screens[ScreenActivityList] = screens.NewActivityList(activityStorage)
    app.screens[ScreenHelp] = screens.NewHelp()
    
    return app
}

func (a *App) Init() tea.Cmd {
    return tea.SetWindowTitle("Fitness TUI")
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        a.width = msg.Width
        a.height = msg.Height
        return a, nil
        
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            if a.currentScreen == ScreenMain {
                return a, tea.Quit
            }
            // Go back to main screen
            a.currentScreen = ScreenMain
            return a, nil
            
        case "h", "?":
            a.currentScreen = ScreenHelp
            return a, nil
            
        case "a":
            if a.currentScreen == ScreenMain {
                a.currentScreen = ScreenActivityList
                return a, a.screens[ScreenActivityList].Init()
            }
            
        case "r":
            if a.currentScreen == ScreenMain {
                a.currentScreen = ScreenRouteList
                return a, nil
            }
            
        case "p":
            if a.currentScreen == ScreenMain {
                a.currentScreen = ScreenPlanList
                return a, nil
            }
        }
    }
    
    // Delegate to current screen
    if screen, exists := a.screens[a.currentScreen]; exists {
        updatedScreen, cmd := screen.Update(msg)
        a.screens[a.currentScreen] = updatedScreen
        return a, cmd
    }
    
    return a, nil
}

func (a *App) View() string {
    if a.currentScreen == ScreenMain {
        return a.renderMainMenu()
    }
    
    if screen, exists := a.screens[a.currentScreen]; exists {
        return screen.View()
    }
    
    return "Unknown screen"
}

func (a *App) renderMainMenu() string {
    style := lipgloss.NewStyle().
        Padding(1, 2).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("62"))
    
    menu := fmt.Sprintf(`Fitness TUI v1.0

Navigation:
[A] Activities - View and analyze your workouts
[R] Routes     - Manage GPX routes and segments  
[P] Plans      - Training plans and workouts
[S] Sync       - Sync with Garmin
[H] Help       - Show help
[Q] Quit       - Exit application

Status: %s`, a.statusMessage)

    return style.Render(menu)
}

func (a *App) Run() error {
    p := tea.NewProgram(a, tea.WithAltScreen())
    _, err := p.Run()
    return err
}
```

### Step 5: Activity List Screen

#### `internal/tui/screens/activity_list.go`
```go
package screens

import (
    "fmt"
    "strings"
    
    "github.com/charmbracelet/bubbles/list"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    
    "github.com/yourusername/fitness-tui/internal/tui/models"
    "github.com/yourusername/fitness-tui/internal/storage"
)

type ActivityList struct {
    list    list.Model
    storage *storage.ActivityStorage
    width   int
    height  int
}

type activityItem struct {
    activity *models.Activity
}

func (i activityItem) FilterValue() string { return i.activity.Name }

func (i activityItem) Title() string {
    return fmt.Sprintf("%s - %s", 
        i.activity.Date.Format("2006-01-02"), 
        i.activity.Name)
}

func (i activityItem) Description() string {
    return fmt.Sprintf("%s  %s  %s", 
        i.activity.FormattedDistance(),
        i.activity.FormattedDuration(),
        i.activity.FormattedPace())
}

func NewActivityList(storage *storage.ActivityStorage) *ActivityList {
    items := []list.Item{}
    
    delegate := list.NewDefaultDelegate()
    delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
        Foreground(lipgloss.Color("170")).
        BorderLeftForeground(lipgloss.Color("170"))
    delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
        Foreground(lipgloss.Color("243"))
    
    l := list.New(items, delegate, 0, 0)
    l.Title = "Activities"
    l.SetShowStatusBar(false)
    l.SetFilteringEnabled(false)
    l.Styles.Title = lipgloss.NewStyle().
        MarginLeft(2).
        Foreground(lipgloss.Color("62")).
        Bold(true)
    
    return &ActivityList{
        list:    l,
        storage: storage,
    }
}

func (m *ActivityList) Init() tea.Cmd {
    return m.loadActivities
}

func (m *ActivityList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.list.SetWidth(msg.Width)
        m.list.SetHeight(msg.Height - 4) // Reserve space for instructions
        return m, nil
        
    case tea.KeyMsg:
        switch msg.String() {
        case "s":
            return m, m.syncActivities
        case "enter":
            if selectedItem := m.list.SelectedItem(); selectedItem != nil {
                // TODO: Navigate to activity detail
                return m, nil
            }
        }
        
    case activitiesLoadedMsg:
        items := make([]list.Item, len(msg.activities))
        for i, activity := range msg.activities {
            items[i] = activityItem{activity: activity}
        }
        m.list.SetItems(items)
        return m, nil
        
    case syncCompleteMsg:
        return m, m.loadActivities
    }
    
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m *ActivityList) View() string {
    instructions := lipgloss.NewStyle().
        Foreground(lipgloss.Color("241")).
        MarginTop(1).
        Render("↑↓ navigate • enter view • s sync • q back")
    
    return fmt.Sprintf("%s\n%s", m.list.View(), instructions)
}

// Commands
type activitiesLoadedMsg struct {
    activities []*models.Activity
}

type syncCompleteMsg struct{}

func (m *ActivityList) loadActivities() tea.Msg {
    activities, err := m.storage.LoadAll()
    if err != nil {
        // TODO: Handle error properly
        return activitiesLoadedMsg{activities: []*models.Activity{}}
    }
    return activitiesLoadedMsg{activities: activities}
}

func (m *ActivityList) syncActivities() tea.Msg {
    // TODO: Implement Garmin sync
    return syncCompleteMsg{}
}
```

### Step 6: ASCII Charts

#### `internal/charts/ascii.go`
```go
package charts

import (
    "fmt"
    "math"
    "strings"
)

const (
    // Unicode block characters for charts
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

// SparklineChart generates a simple sparkline from data points
func SparklineChart(data []float64, width int) string {
    if len(data) == 0 {
        return strings.Repeat(BlockEmpty, width)
    }
    
    // Find min/max for normalization
    min, max := data[0], data[0]
    for _, v := range data {
        if v < min {
            min = v
        }
        if v > max {
            max = v
        }
    }
    
    // Handle case where all values are the same
    if min == max {
        return strings.Repeat(Block4, width)
    }
    
    // Downsample data to fit width
    sampledData := sampleData(data, width)
    
    var result strings.Builder
    for _, value := range sampledData {
        // Normalize to 0-8 range (9 levels including empty)
        normalized := (value - min) / (max - min)
        level := int(normalized * 8)
        if level > 8 {
            level = 8
        }
        result.WriteString(blockChars[level])
    }
    
    return result.String()
}

// LineChart generates a multi-line ASCII chart
func LineChart(data []float64, width, height int, title string) string {
    if len(data) == 0 {
        return fmt.Sprintf("%s\n%s", title, strings.Repeat("-", width))
    }
    
    // Find min/max for scaling
    min, max := data[0], data[0]
    for _, v := range data {
        if v < min {
            min = v
        }
        if v > max {
            max = v
        }
    }
    
    if min == max {
        // All values are the same
        line := strings.Repeat("-", width)
        var chart strings.Builder
        chart.WriteString(fmt.Sprintf("%s (%.2f)\n", title, min))
        for i := 0; i < height; i++ {
            if i == height/2 {
                chart.WriteString(line + "\n")
            } else {
                chart.WriteString(strings.Repeat(" ", width) + "\n")
            }
        }
        return chart.String()
    }
    
    // Sample data to fit width
    sampledData := sampleData(data, width)
    
    // Create chart grid
    grid := make([][]rune, height)
    for i := range grid {
        grid[i] = make([]rune, width)
        for j := range grid[i] {
            grid[i][j] = ' '
        }
    }
    
    // Plot data points
    for x, value := range sampledData {
        // Scale to chart height (inverted because we draw top to bottom)
        normalized := (value - min) / (max - min)
        y := height - 1 - int(normalized*float64(height-1))
        if y < 0 {
            y = 0
        }
        if y >= height {
            y = height - 1
        }
        
        if x < width {
            grid[y][x] = '•'
        }
    }
    
    // Render chart
    var chart strings.Builder
    chart.WriteString(fmt.Sprintf("%s (%.2f - %.2f)\n", title, min, max))
    
    for _, row := range grid {
        chart.WriteString(string(row) + "\n")
    }
    
    return chart.String()
}

// sampleData downsamples data array to target length
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

// ChartData represents different types of chart data
type ChartData struct {
    Label  string
    Values []float64
    Unit   string
}

// MultiChart renders multiple charts in a stacked view
func MultiChart(charts []ChartData, width int) string {
    var result strings.Builder
    
    for i, chart := range charts {
        if i > 0 {
            result.WriteString("\n")
        }
        
        title := chart.Label
        if chart.Unit != "" {
            title += fmt.Sprintf(" (%s)", chart.Unit)
        }
        
        sparkline := SparklineChart(chart.Values, width)
        result.WriteString(fmt.Sprintf("%-12s %s", title+":", sparkline))
    }
    
    return result.String()
}
```

### Step 7: Main Entry Point

#### `cmd/main.go`
```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "github.com/yourusername/fitness-tui/internal/config"
    "github.com/yourusername/fitness-tui/internal/tui"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    
    // Validate required configuration
    if err := validateConfig(cfg); err != nil {
        fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
        fmt.Fprintf(os.Stderr, "Please edit ~/.fitness-tui/config.yaml and add your credentials\n")
        os.Exit(1)
    }
    
    // Create and run TUI application
    app := tui.NewApp(cfg)
    if err := app.Run(); err != nil {
        log.Fatalf("Application error: %v", err)
    }
}

func validateConfig(cfg *config.Config) error {
    if cfg.Garmin.Username == "" {
        return fmt.Errorf("Garmin username is required")
    }
    
    if cfg.Garmin.Password == "" {
        return fmt.Errorf("Garmin password is required")
    }
    
    if cfg.OpenRouter.APIKey == "" {
        return fmt.Errorf("OpenRouter API key is required for analysis features")
    }
    
    return nil
}
```

## Testing Strategy

### Unit Tests Structure
```bash
# Create test files
touch internal/storage/activities_test.go
touch internal/charts/ascii_test.go
touch internal/tui/models/activity_test.go
```

#### Example Test File: `internal/charts/ascii_test.go`
```go
package charts

import (
    "strings"
    "testing"
    
    "github.com/stretchr/testify/assert"
)

func TestSparklineChart(t *testing.T) {
    tests := []struct {
        name     string
        data     []float64
        width    int
        expected string
    }{
        {
            name:     "empty data",
            data:     []float64{},
            width:    10,
            expected: "          ", // 10 spaces
        },
        {
            name:     "single value",
            data:     []float64{50},
            width:    5,
            expected: "▄▄▄▄▄",
        },
        {
            name:     "ascending values",
            data:     []float64{1, 2, 3, 4, 5},
            width:    5,
            expected: "▁▂▄▆█",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := SparklineChart(tt.data, tt.width)
            assert.Equal(t, tt.expected, result)
            assert.Len(t, result, tt.width)
        })
    }
}

func TestLineChart(t *testing.T) {
    data := []float64{1, 3, 2, 5, 4}
    result := LineChart(data, 10, 5, "Test Chart")
    
    // Check that it returns a multi-line string
    lines := strings.Split(result, "\n")
    assert.True(t, len(lines) >= 5)
    
    // Check title is included
    assert.Contains(t, lines[0], "Test Chart")
}
```

### Integration Tests

#### `test/integration/storage_test.go`
```go
package integration

import (
    "os"
    "path/filepath"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "github.com/yourusername/fitness-tui/internal/storage"
    "github.com/yourusername/fitness-tui/internal/tui/models"
)

func TestActivityStorageIntegration(t *testing.T) {
    // Create temporary directory
    tmpDir, err := os.MkdirTemp("", "fitness-tui-test-*")
    require.NoError(t, err)
    defer os.RemoveAll(tmpDir)
    
    storage := storage.NewActivityStorage(tmpDir)
    
    // Create test activity
    activity := &models.Activity{
        ID:       "test-123",
        Name:     "Test Run",
        Type:     "running",
        Date:     time.Now(),
        Duration: 30 * time.Minute,
        Distance: 5000, // 5km
        Metrics: models.ActivityMetrics{
            AvgHeartRate: 150,
            AvgPace:      300, // 5:00/km
        },
    }
    
    // Test save
    err = storage.Save(activity)
    require.NoError(t, err)
    
    // Test load
    activities, err := storage.LoadAll()
    require.NoError(t, err)
    require.Len(t, activities, 1)
    
    loaded := activities[0]
    assert.Equal(t, activity.ID, loaded.ID)
    assert.Equal(t, activity.Name, loaded.Name)
    assert.Equal(t, activity.Distance, loaded.Distance)
}
```

### Makefile for Development

#### `Makefile`
```makefile
.PHONY: build test run clean install dev lint fmt

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=fitness-tui

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run the application
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd
	./$(BINARY_NAME)

# Development mode with hot reload
dev:
	air

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Install dependencies
install:
	$(GOMOD) tidy
	$(GOMOD) download

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	goimports -w .
	$(GOCMD) fmt ./...

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd

# Install the binary
install-bin: build
	cp $(BINARY_NAME) $(GOPATH)/bin/
```

### Step 8: Garmin Integration

#### `internal/garmin/client.go`
```go
package garmin

import (
    "context"
    "fmt"
    "time"
    
    "github.com/sstent/go-garth"
    
    "github.com/yourusername/fitness-tui/internal/tui/models"
)

type Client struct {
    client   *garth.Client
    username string
    password string
}

func NewClient(username, password string) *Client {
    return &Client{
        username: username,
        password: password,
    }
}

func (c *Client) Connect() error {
    client := garth.NewClient(garth.Credentials{
        Username: c.username,
        Password: c.password,
    })
    
    if err := client.Login(); err != nil {
        return fmt.Errorf("failed to login to Garmin: %w", err)
    }
    
    c.client = client
    return nil
}

func (c *Client) GetActivities(ctx context.Context, limit int) ([]*models.Activity, error) {
    if c.client == nil {
        return nil, fmt.Errorf("client not connected")
    }
    
    // Get activities from Garmin
    gActivities, err := c.client.GetActivities(ctx, 0, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch activities: %w", err)
    }
    
    var activities []*models.Activity
    for _, gActivity := range gActivities {
        activity := c.convertActivity(gActivity)
        activities = append(activities, activity)
    }
    
    return activities, nil
}

func (c *Client) GetActivityDetails(ctx context.Context, activityID string) (*models.Activity, error) {
    if c.client == nil {
        return nil, fmt.Errorf("client not connected")
    }
    
    details, err := c.client.GetActivityDetails(ctx, activityID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch activity details: %w", err)
    }
    
    activity := c.convertActivityWithDetails(details)
    return activity, nil
}

func (c *Client) convertActivity(gActivity interface{}) *models.Activity {
    // This function needs to be implemented based on the actual
    // structure returned by go-garth library
    // The exact implementation depends on the go-garth API
    
    // Placeholder implementation
    return &models.Activity{
        ID:       fmt.Sprintf("garmin-%v", gActivity),
        Name:     "Imported Activity",
        Type:     "running",
        Date:     time.Now(),
        Duration: 30 * time.Minute,
        Distance: 5000,
        Metrics:  models.ActivityMetrics{},
    }
}

func (c *Client) convertActivityWithDetails(details interface{}) *models.Activity {
    // Similar to convertActivity but with more detailed information
    // Implementation depends on go-garth API structure
    
    activity := c.convertActivity(details)
    // Add detailed metrics here
    return activity
}
```

### Step 9: LLM Analysis Integration

#### `internal/analysis/llm.go`
```go
package analysis

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/yourusername/fitness-tui/internal/tui/models"
)

type OpenRouterClient struct {
    apiKey  string
    baseURL string
    model   string
    client  *http.Client
}

type OpenRouterRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
    MaxTokens int      `json:"max_tokens,omitempty"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type OpenRouterResponse struct {
    Choices []Choice `json:"choices"`
    Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
    Message Message `json:"message"`
}

type APIError struct {
    Message string `json:"message"`
    Type    string `json:"type"`
}

func NewOpenRouterClient(apiKey, baseURL, model string) *OpenRouterClient {
    return &OpenRouterClient{
        apiKey:  apiKey,
        baseURL: baseURL,
        model:   model,
        client:  &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *OpenRouterClient) AnalyzeActivity(ctx context.Context, activity *models.Activity, workoutContext string) (*models.Analysis, error) {
    prompt := c.buildAnalysisPrompt(activity, workoutContext)
    
    request := OpenRouterRequest{
        Model: c.model,
        Messages: []Message{
            {
                Role:    "user",
                Content: prompt,
            },
        },
        MaxTokens: 1000,
    }
    
    requestBody, err := json.Marshal(request)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }
    
    req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.apiKey)
    
    resp, err := c.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to make request: %w", err)
    }
    defer resp.Body.Close()
    
    var response OpenRouterResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    if response.Error != nil {
        return nil, fmt.Errorf("API error: %s", response.Error.Message)
    }
    
    if len(response.Choices) == 0 {
        return nil, fmt.Errorf("no response generated")
    }
    
    content := response.Choices[0].Message.Content
    insights := c.extractInsights(content)
    
    analysis := &models.Analysis{
        ActivityID:  activity.ID,
        WorkoutType: workoutContext,
        GeneratedAt: time.Now(),
        Content:     content,
        Insights:    insights,
    }
    
    return analysis, nil
}

func (c *OpenRouterClient) buildAnalysisPrompt(activity *models.Activity, workoutContext string) string {
    return fmt.Sprintf(`Analyze this fitness activity in the context of the intended workout:

INTENDED WORKOUT: %s

ACTIVITY DETAILS:
- Name: %s
- Type: %s  
- Date: %s
- Duration: %s
- Distance: %s
- Average Heart Rate: %d bpm
- Average Pace: %s
- Elevation Gain: %.0fm

Please provide:
1. How well did this activity match the intended workout?
2. Performance analysis (pacing, heart rate zones, effort level)
3. Areas for improvement
4. Recovery recommendations
5. Any red flags or injury risk factors

Format your response in markdown with clear sections.`, 
        workoutContext,
        activity.Name,
        activity.Type,
        activity.Date.Format("2006-01-02"),
        activity.FormattedDuration(),
        activity.FormattedDistance(),
        activity.Metrics.AvgHeartRate,
        activity.FormattedPace(),
        activity.Metrics.ElevationGain,
    )
}

func (c *OpenRouterClient) extractInsights(content string) []string {
    // Simple insight extraction - look for bullet points or numbered lists
    // This is a basic implementation - could be enhanced with NLP
    
    insights := []string{}
    lines := strings.Split(content, "\n")
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
            insight := strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
            if len(insight) > 10 { // Filter out very short lines
                insights = append(insights, insight)
            }
        }
    }
    
    return insights
}
```

#### `internal/analysis/cache.go`
```go
package analysis

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"
    
    "github.com/yourusername/fitness-tui/internal/tui/models"
)

type AnalysisCache struct {
    dataDir string
}

func NewAnalysisCache(dataDir string) *AnalysisCache {
    analysisDir := filepath.Join(dataDir, "analysis")
    os.MkdirAll(analysisDir, 0755)
    
    return &AnalysisCache{
        dataDir: dataDir,
    }
}

func (c *AnalysisCache) Save(analysis *models.Analysis) error {
    // Save as markdown file for easy reading
    mdPath := c.getMarkdownPath(analysis.ActivityID, analysis.WorkoutType)
    if err := os.WriteFile(mdPath, []byte(analysis.Content), 0644); err != nil {
        return fmt.Errorf("failed to save analysis markdown: %w", err)
    }
    
    // Save metadata as JSON
    metaPath := c.getMetadataPath(analysis.ActivityID, analysis.WorkoutType)
    metadata := struct {
        ActivityID   string    `json:"activity_id"`
        WorkoutType  string    `json:"workout_type"`
        GeneratedAt  time.Time `json:"generated_at"`
        Insights     []string  `json:"insights"`
        MarkdownPath string    `json:"markdown_path"`
    }{
        ActivityID:   analysis.ActivityID,
        WorkoutType:  analysis.WorkoutType,
        GeneratedAt:  analysis.GeneratedAt,
        Insights:     analysis.Insights,
        MarkdownPath: mdPath,
    }
    
    data, err := json.MarshalIndent(metadata, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal analysis metadata: %w", err)
    }
    
    if err := os.WriteFile(metaPath, data, 0644); err != nil {
        return fmt.Errorf("failed to save analysis metadata: %w", err)
    }
    
    analysis.FilePath = mdPath
    return nil
}

func (c *AnalysisCache) Load(activityID, workoutType string) (*models.Analysis, error) {
    metaPath := c.getMetadataPath(activityID, workoutType)
    
    if _, err := os.Stat(metaPath); os.IsNotExist(err) {
        return nil, nil // No cached analysis
    }
    
    data, err := os.ReadFile(metaPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read analysis metadata: %w", err)
    }
    
    var metadata struct {
        ActivityID   string    `json:"activity_id"`
        WorkoutType  string    `json:"workout_type"`
        GeneratedAt  time.Time `json:"generated_at"`
        Insights     []string  `json:"insights"`
        MarkdownPath string    `json:"markdown_path"`
    }
    
    if err := json.Unmarshal(data, &metadata); err != nil {
        return nil, fmt.Errorf("failed to unmarshal analysis metadata: %w", err)
    }
    
    // Load markdown content
    content, err := os.ReadFile(metadata.MarkdownPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read analysis content: %w", err)
    }
    
    analysis := &models.Analysis{
        ActivityID:  metadata.ActivityID,
        WorkoutType: metadata.WorkoutType,
        GeneratedAt: metadata.GeneratedAt,
        Content:     string(content),
        Insights:    metadata.Insights,
        FilePath:    metadata.MarkdownPath,
    }
    
    return analysis, nil
}

func (c *AnalysisCache) getMarkdownPath(activityID, workoutType string) string {
    filename := fmt.Sprintf("%s-%s-analysis.md", activityID, sanitizeFilename(workoutType))
    return filepath.Join(c.dataDir, "analysis", filename)
}

func (c *AnalysisCache) getMetadataPath(activityID, workoutType string) string {
    filename := fmt.Sprintf("%s-%s-meta.json", activityID, sanitizeFilename(workoutType))
    return filepath.Join(c.dataDir, "analysis", filename)
}

func sanitizeFilename(name string) string {
    // Replace invalid filename characters
    replacer := strings.NewReplacer(
        "/", "-", "\\", "-", ":", "-", "*", "-",
        "?", "-", "\"", "-", "<", "-", ">", "-",
        "|", "-", " ", "-",
    )
    return replacer.Replace(name)
}
```

### Step 10: Development Configuration

#### `.air.toml` (for hot reload during development)
```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd"
  delay = 0
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "dist"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

#### `examples/sample-config.yaml`
```yaml
garmin:
  username: "your-garmin-username"
  password: "your-garmin-password"

openrouter:
  api_key: "your-openrouter-api-key"
  model: "deepseek/deepseek-r1-05028:free"
  base_url: "https://openrouter.ai/api/v1"

storage:
  data_dir: "~/.fitness-tui"
```

## Deployment

### Building for Release

#### `.goreleaser.yaml`
```yaml
project_name: fitness-tui

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - env:
      - CGO_ENABLED=0
    main: ./cmd
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- title .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else if eq .Arch "386" }}i386
      {{- else }}{{ .Arch }}{{ end }}
      {{- if .Arm }}v{{ .Arm }}{{ end }}
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'

release:
  github:
    owner: yourusername
    name: fitness-tui
```

### Installation Script

#### `install.sh`
```bash
#!/bin/bash
set -e

# Fitness TUI Installation Script

REPO="yourusername/fitness-tui"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="x86_64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

echo "Installing Fitness TUI for $OS-$ARCH..."

# Get latest release
LATEST=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep -o '"tag_name": "[^"]*' | cut -d'"' -f4)

if [ -z "$LATEST" ]; then
    echo "Failed to get latest release"
    exit 1
fi

echo "Latest version: $LATEST"

# Download URL
FILENAME="fitness-tui_${OS^}_${ARCH}.tar.gz"
if [ "$OS" = "windows" ]; then
    FILENAME="fitness-tui_Windows_${ARCH}.zip"
fi

URL="https://github.com/$REPO/releases/download/$LATEST/$FILENAME"

echo "Downloading $URL..."
curl -L "$URL" -o "/tmp/$FILENAME"

# Extract and install
cd /tmp
if [ "$OS" = "windows" ]; then
    unzip -q "$FILENAME"
else
    tar -xzf "$FILENAME"
fi

# Move to install directory
sudo mv fitness-tui "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/fitness-tui"

echo "Fitness TUI installed successfully to $INSTALL_DIR/fitness-tui"
echo "Run 'fitness-tui' to get started!"

# Cleanup
rm -f "/tmp/$FILENAME"
```

## Development Workflow

### Daily Development Commands
```bash
# Start development mode
make dev

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Lint code
make lint

# Build application
make build

# Run the built application
./fitness-tui
```

### Git Workflow
```bash
# Feature development
git checkout -b feature/activity-analysis
# ... make changes ...
make fmt lint test
git add .
git commit -m "feat: add activity analysis with LLM integration"
git push origin feature/activity-analysis
# Create PR

# Release preparation
git checkout main
git tag v1.0.0
git push origin v1.0.0
# GitHub Actions will build and release
```

## Next Steps

1. **Phase 1 Implementation**: Start with the core foundation
   - Set up project structure
   - Implement configuration system
   - Create basic TUI with bubbletea
   - Add file storage for activities

2. **Testing**: Write tests as you implement each component
   - Unit tests for business logic
   - Integration tests for storage
   - TUI interaction tests

3. **Documentation**: 
   - Add godoc comments to all public functions
   - Create user documentation
   - Add examples and troubleshooting guides

4. **Performance**: 
   - Profile the application
   - Optimize chart rendering
   - Add caching where appropriate

This implementation guide provides a complete roadmap for building the fitness TUI application. Each step includes detailed code examples and follows Go best practices for maintainable, testable code.