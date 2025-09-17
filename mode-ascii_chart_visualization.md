# ASCII Chart Visualization Mode Rules

## Core Principles
- Charts must render efficiently with minimal terminal flickering
- Support datasets up to 10,000 points with smooth downsampling
- Responsive to terminal resize events
- Consistent styling with lipgloss
- Graceful degradation for terminals with limited width/height

## Implementation Guidelines

### Data Processing
1. Downsample large datasets using Largest-Triangle-Three-Buckets algorithm
2. Normalize values to terminal height minus chart margins
3. Handle missing data points gracefully (interpolate or skip)

### Rendering
- Use Unicode block characters (▁▂▃▄▅▆▇█) for vertical resolution
- Implement horizontal scaling based on terminal width
- Add axis labels with metric units and scale indicators
- Color-code different metrics:
  - Heart Rate: Red
  - Power: Yellow
  - Elevation: Cyan

### Performance Optimization
- Pre-calculate downsampled data when possible
- Limit re-rendering to when data changes or terminal resizes
- Use efficient string building with strings.Builder
- Implement rendering caching for static charts

### Error Handling
- Display "No Data" message when dataset is empty
- Show "Rendering Error" with details in debug mode
- Handle terminal size too small with informative message

### Testing Requirements
- 90% test coverage for chart logic
- Golden tests for various chart configurations
- Performance benchmarks for downsampling and rendering
- Edge case tests: single point, empty data, negative values

## Required Components
1. Downsampling module (`internal/types/downsample.go`)
2. Chart component (`internal/tui/components/chart.go`)
3. Chart model for state management
4. Integration with activity detail screen (`internal/tui/screens/activity_detail.go`)

## Integration Example
```go
// In activity_detail.go
func (m *ActivityDetailModel) renderCharts() string {
    hrChart := components.NewChart(m.activity.HRData, "HR (bpm)")
    powerChart := components.NewChart(m.activity.PowerData, "Power (w)")
    elevationChart := components.NewChart(m.activity.ElevationData, "Elevation (m)")
    
    return lipgloss.JoinVertical(
        lipgloss.Left,
        hrChart.View(),
        powerChart.View(),
        elevationChart.View(),
    )
}
```

## Dependencies
- Requires populated activity data model
- Relies on downsampling functionality
- Uses terminal dimensions from bubbletea

## Quality Standards
- Zero memory leaks
- Handle datasets up to 10,000 points efficiently
- Responsive rendering (<100ms for full redraw)
- Accessible color schemes
- Comprehensive test coverage