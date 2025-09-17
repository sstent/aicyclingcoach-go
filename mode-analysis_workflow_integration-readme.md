# Analysis Workflow Integration Mode

## Purpose
This mode focuses on integrating the AI analysis generation workflow into the application, connecting the OpenRouter client with the activity detail screen and caching system. The integration must handle both cached and real-time analysis scenarios.

## Key Features
- Seamless triggering of analysis from activity detail screen
- Automatic retrieval of cached analysis when available
- Background generation of new analysis with progress feedback
- Integration with storage layer for caching results
- Error handling for analysis failures

## Integration Points
- [`internal/tui/screens/activity_detail.go`](fitness-tui/internal/tui/screens/activity_detail.go) - Analysis triggering
- [`internal/analysis/service.go`](fitness-tui/internal/analysis/service.go) - Analysis service
- [`internal/storage/analysis.go`](fitness-tui/internal/storage/analysis.go) - Analysis caching
- [`internal/tui/app.go`](fitness-tui/internal/tui/app.go) - Background task management

## Dependencies
- Requires fully implemented OpenRouter client
- Relies on analysis caching system
- Needs activity data model population

## Expected Outcomes
- Seamless analysis generation from activity detail screen
- Efficient use of cached analysis when available
- Background processing with user feedback
- Comprehensive error handling for analysis failures