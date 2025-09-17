# Analysis Workflow Integration Mode Rules

## Core Principles
- Seamless integration between UI, analysis service, and caching layer
- Prioritize cached analysis when available for instant feedback
- Background processing for new analysis with non-blocking UI
- Comprehensive error handling for API failures and edge cases
- Clear user feedback during analysis generation

## Implementation Guidelines

### Workflow Triggers
1. Implement analysis trigger in [`activity_detail.go`](fitness-tui/internal/tui/screens/activity_detail.go) via key binding (e.g., 'a')
2. Check analysis cache before initiating new requests
3. Show immediate feedback when analysis starts (spinner + status)

### Cached Analysis Handling
- Query cache using activity ID and analysis type
- Display cached analysis immediately if available
- Include freshness indicator (e.g., "Cached 2h ago")
- Provide option to refresh cached analysis

### Real-time Analysis Generation
1. Prepare analysis context:
   - Activity metrics
   - User preferences
   - Training plan context
2. Use prompt templates from [`prompts.go`](fitness-tui/internal/analysis/prompts.go)
3. Send request via OpenRouter client with timeout
4. Parse and validate response
5. Cache results (markdown + metadata)

### Background Processing
- Implement as bubbletea Command
- Use goroutine with proper context cancellation
- Send progress messages to UI
- Handle termination on app exit

### Error Handling
- **Cache Errors**: Log but proceed with analysis
- **API Errors**: 
  - Rate limits: Show retry timer
  - Authentication: Prompt to check credentials
  - Timeouts: Offer to retry
- **Parsing Errors**: Fallback to raw response

### User Feedback
- **States**:
  - "Loading cached analysis..."
  - "Generating analysis..."
  - "Analysis complete!"
  - "Analysis failed: [reason]"
- **Progress Indicators**:
  - Spinner for indeterminate progress
  - Percent complete for large analyses
- **Success Metrics**: Display token usage/cost

## Integration Points
1. Activity Detail Screen ([`activity_detail.go`](fitness-tui/internal/tui/screens/activity_detail.go)):
   - Add analysis trigger key binding
   - Implement analysis status display
   - Handle analysis messages

2. Analysis Service ([`service.go`](fitness-tui/internal/analysis/service.go)):
   - Coordinate cache check → API request → caching
   - Handle error states and retries
   - Format analysis for display

3. App Model ([`app.go`](fitness-tui/internal/tui/app.go)):
   - Manage background command lifecycle
   - Handle cross-screen messaging
   - Global error handling

## Quality Standards
- Zero UI blocking during analysis
- Comprehensive test coverage (>85%)
- Graceful degradation when APIs unavailable
- Sensible timeouts (30s API, 10s cache)
- Secure credential handling
- Efficient memory usage during large analyses