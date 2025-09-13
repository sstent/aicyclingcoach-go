package garmin

import (
	"context"
)

// Sync performs the complete synchronization process
func (c *Client) Sync(ctx context.Context, logger Logger) (int, error) {
	// Authenticate
	if err := c.Connect(logger); err != nil {
		return 0, err
	}

	// Get activities
	activities, err := c.GetActivities(ctx, 50, logger)
	if err != nil {
		return 0, err
	}

	return len(activities), nil
}
