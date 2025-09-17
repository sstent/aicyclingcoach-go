# ASCII Chart Visualization Mode

## Purpose
This mode focuses on implementing terminal-based charts for activity metrics (HR, power, elevation) using Unicode block characters. The charts must be responsive to terminal size and efficiently render large datasets.

## Key Features
- Data downsampling for large datasets
- Normalization of values to terminal height
- Rendering with Unicode block characters (▁▂▃▄▅▆▇█)
- Responsive design that adapts to terminal resize events
- Integration with the activity detail screen

## Integration Points
- [`internal/tui/components/chart.go`](fitness-tui/internal/tui/components/chart.go)
- [`internal/types/downsample.go`](fitness-tui/internal/types/downsample.go)
- [`internal/tui/screens/activity_detail.go`](fitness-tui/internal/tui/screens/activity_detail.go)

## Dependencies
- Requires activity data model to be populated
- Relies on downsampling functionality

## Expected Outcomes
- Charts for HR, power, and elevation in activity detail view
- Efficient rendering of up to 10,000 data points
- Unit tests covering 90% of chart logic