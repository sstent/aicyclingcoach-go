package garmin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sstent/fitness-tui/internal/garmin/garth"
	"github.com/sstent/fitness-tui/internal/garmin/garth/client"
	"github.com/sstent/fitness-tui/internal/tui/models"
)

type GarminClient interface {
	Connect(logger Logger) error
	GetActivities(ctx context.Context, limit int, logger Logger) ([]*models.Activity, error)
}

type Client struct {
	username    string
	password    string
	storagePath string
	garthClient *client.Client
}

func NewClient(username, password, storagePath string) *Client {
	return &Client{
		username:    username,
		password:    password,
		storagePath: storagePath,
	}
}

func (c *Client) Connect(logger Logger) error {
	if logger == nil {
		logger = &NoopLogger{}
	}
	logger.Infof("Starting Garmin authentication")

	// Create client with default domain
	garthClient, err := garth.NewClient("garmin.com")
	if err != nil {
		logger.Errorf("Failed to create Garmin client: %v", err)
		return err
	}
	c.garthClient = garthClient

	// Check for existing session
	sessionFile := filepath.Join(c.storagePath, "garmin_session.json")
	if _, err := os.Stat(sessionFile); err == nil {
		if err := c.garthClient.LoadSession(sessionFile); err == nil {
			logger.Infof("Loaded existing Garmin session")
			return nil
		}
	}

	// Perform login
	if err := c.garthClient.Login(c.username, c.password); err != nil {
		logger.Errorf("Garmin authentication failed: %v", err)
		return err
	}

	// Save session for future use
	if err := c.garthClient.SaveSession(sessionFile); err != nil {
		logger.Warnf("Failed to save Garmin session: %v", err)
	}

	logger.Infof("Authentication successful")
	return nil
}

func (c *Client) GetActivities(ctx context.Context, limit int, logger Logger) ([]*models.Activity, error) {
	if logger == nil {
		logger = &NoopLogger{}
	}
	logger.Infof("Fetching %d activities from Garmin Connect", limit)

	if c.garthClient == nil {
		if err := c.Connect(logger); err != nil {
			return nil, err
		}
	}

	// Get activities from Garmin API
	garthActivities, err := c.garthClient.GetActivities(limit)
	if err != nil {
		logger.Errorf("Failed to fetch activities: %v", err)
		return nil, err
	}

	// Convert to our internal model
	activities := make([]*models.Activity, 0, len(garthActivities))
	for _, ga := range garthActivities {
		// Use the already parsed time from CustomTime struct
		if ga.StartTimeGMT.IsZero() {
			logger.Warnf("Activity %d has invalid start time", ga.ActivityID)
			continue
		}

		activities = append(activities, &models.Activity{
			ID:        fmt.Sprintf("%d", ga.ActivityID),
			Name:      ga.ActivityName,
			Type:      ga.ActivityType.TypeKey,
			Date:      ga.StartTimeGMT.Time, // Access the parsed time directly
			Distance:  ga.Distance,
			Duration:  time.Duration(ga.Duration) * time.Second,
			Elevation: ga.ElevationGain,
			Calories:  int(ga.Calories),
		})
	}

	logger.Infof("Successfully fetched %d activities", len(activities))
	return activities, nil
}
