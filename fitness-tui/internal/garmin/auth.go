package garmin

import (
	"fmt"

	"github.com/sstent/fitness-tui/internal/garmin/garth"
)

// Authenticate performs Garmin Connect authentication
func (c *Client) Authenticate(logger Logger) error {
	logger.Infof("Authenticating with username: %s", c.username)

	// Initialize Garth client
	garthClient := garth.New()

	// Perform authentication
	if err := garthClient.Authenticate(c.username, c.password); err != nil {
		logger.Errorf("Authentication failed: %v", err)
		return fmt.Errorf("authentication failed: %w", err)
	}

	logger.Infof("Authentication successful")
	return nil
}
