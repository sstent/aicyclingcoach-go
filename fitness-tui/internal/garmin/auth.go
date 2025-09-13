package garmin

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sstent/fitness-tui/internal/garmin/garth/client"
)

// Authenticate performs Garmin Connect authentication
func (c *Client) Authenticate(logger Logger) error {
	logger.Infof("Authenticating with username: %s", c.username)

	// Initialize Garth client
	garthClient, err := client.NewClient("garmin.com")
	if err != nil {
		logger.Errorf("Failed to create Garmin client: %v", err)
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Try to load existing session
	sessionFile := filepath.Join(os.Getenv("HOME"), ".fitness-tui", "garmin_session.json")
	if err := garthClient.LoadSession(sessionFile); err != nil {
		logger.Infof("No existing session found, logging in with credentials")

		// Perform authentication if no session exists
		if err := garthClient.Login(c.username, c.password); err != nil {
			logger.Errorf("Authentication failed: %v", err)
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Save session for future use
		if err := garthClient.SaveSession(sessionFile); err != nil {
			logger.Warnf("Failed to save session: %v", err)
		}
	} else {
		logger.Infof("Loaded existing session")
	}

	// Store the authenticated client
	c.garthClient = garthClient

	logger.Infof("Authentication successful")
	return nil
}
