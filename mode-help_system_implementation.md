# Help System Implementation Mode Rules

## Core Principles
- Help system must be accessible from any screen with a consistent shortcut ('?')
- Context-sensitive content based on active screen
- Keyboard-first navigation within the help system
- Responsive layout adapting to terminal size
- Minimal performance impact on application

## Implementation Guidelines

### Global Key Handling
1. Register global '?' key handler in [`app.go`](fitness-tui/internal/tui/app.go)
2. Toggle help screen regardless of current active screen
3. Preserve current screen state when help is active
4. Implement help stack to support nested help contexts

### Help Screen Components
- **Main Help Screen** ([`screens/help.go`](fitness-tui/internal/tui/screens/help.go)):
  - Display context-sensitive help based on previous screen
  - Show keyboard shortcuts for current context
  - Provide navigation to general help topics
  
- **Help Panel Component** ([`components/help_panel.go`](fitness-tui/internal/tui/components/help_panel.go)):
  - Reusable component for displaying help content
  - Support scrolling for longer content
  - Style with lipgloss for consistent look

### Context-Sensitive Content
- Define help content for each screen:
  - Activity List: Navigation, syncing, selection
  - Activity Detail: Analysis, chart viewing, navigation
  - Route List: GPX management, route selection
- Store help content in structured format (YAML/JSON)
- Load content dynamically based on context

### Navigation System
- Tab-based navigation between help sections
- Keyboard shortcuts for:
  - Close help ('esc' or 'q')
  - Navigate sections (tab/shift-tab)
  - Scroll content (arrow keys)
- Breadcrumb trail for hierarchical help

### Performance Optimization
- Preload help content during app initialization
- Cache rendered help panels
- Lazy load detailed topic content
- Minimize re-renders during navigation

### Error Handling
- Fallback content for missing help definitions
- Graceful degradation on rendering errors
- Help system available even when other features fail

### Testing Requirements
- 100% test coverage for help navigation logic
- Golden tests for help content rendering
- Integration tests for context switching
- Test help availability during error states

## Integration Points
1. Global key handler in app model ([`app.go`](fitness-tui/internal/tui/app.go))
2. Screen-specific help content definitions
3. Help panel integration in each screen
4. Help content storage and loading

## Quality Standards
- Instant help toggle (<100ms response time)
- Consistent styling across all screens
- Comprehensive coverage of all features
- Accessible navigation without mouse
- Zero performance impact on main application