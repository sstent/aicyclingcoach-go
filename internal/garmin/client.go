package garmin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sony/gobreaker"
	"github.com/sstent/fitness-tui/internal/garmin/garth"
	"github.com/sstent/fitness-tui/internal/garmin/garth/client"
	"github.com/sstent/fitness-tui/internal/tui/models"
)

type GarminClient interface {
	Connect(logger Logger) error
	GetActivities(ctx context.Context, limit int, logger Logger) ([]*models.Activity, error)
	GetAllActivities(ctx context.Context, logger Logger) ([]models.Activity, error)
	DownloadActivityFile(ctx context.Context, activityID string, format string, logger Logger) ([]byte, error)
}

type Client struct {
	username    string
	password    string
	storagePath string
	garthClient *client.Client
	cb          *gobreaker.CircuitBreaker
}

func NewClient(username, password, storagePath string) *Client {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "GarminClient",
		MaxRequests: 1,
		Interval:    0,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	})
	return &Client{
		username:    username,
		password:    password,
		storagePath: storagePath,
		cb:          cb,
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
		return &AuthenticationError{Err: err}
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

	// Wrap API call with circuit breaker
	resp, err := c.cb.Execute(func() (interface{}, error) {
		return c.garthClient.GetActivities(limit, 0)
	})
	if err != nil {
		logger.Errorf("Failed to fetch activities (circuit breaker): %v", err)
		return nil, err
	}
	garthActivities := resp.([]*garth.Activity)

	// Convert to our internal model
	activities := make([]*models.Activity, 0, len(garthActivities))
	for _, ga := range garthActivities {
		// Use the already parsed time from CustomTime struct
		if ga.StartTimeGMT.IsZero() {
			logger.Warnf("Activity %d has invalid start time", ga.ActivityID)
			continue
		}

		// Convert garth activity to internal model
		activity := &models.Activity{
			ID:        fmt.Sprintf("%d", ga.ActivityID),
			Name:      ga.ActivityName,
			Type:      ga.ActivityType.TypeKey,
			Date:      ga.StartTimeGMT.Time, // Access the parsed time directly
			Distance:  ga.Distance,
			Duration:  time.Duration(ga.Duration) * time.Second,
			Elevation: ga.ElevationGain,
			Calories:  int(ga.Calories),
		}

		// Populate metrics from garth data
		if ga.AverageHR > 0 {
			activity.Metrics.AvgHeartRate = int(ga.AverageHR)
		}
		if ga.MaxHR > 0 {
			activity.Metrics.MaxHeartRate = int(ga.MaxHR)
		}
		if ga.AverageSpeed > 0 {
			// Convert m/s to km/h
			activity.Metrics.AvgSpeed = ga.AverageSpeed * 3.6
		}
		if ga.ElevationGain > 0 {
			activity.Metrics.ElevationGain = ga.ElevationGain
		}
		if ga.ElevationLoss > 0 {
			activity.Metrics.ElevationLoss = ga.ElevationLoss
		}
		if ga.AverageSpeed > 0 && ga.Distance > 0 {
			// Calculate pace: seconds per km
			activity.Metrics.AvgPace = (ga.Duration / ga.Distance) * 1000
		}

		activities = append(activities, activity)
	}

	logger.Infof("Successfully fetched %d activities", len(activities))
	return activities, nil
}

func (c *Client) GetAllActivities(ctx context.Context, logger Logger) ([]models.Activity, error) {
	if logger == nil {
		logger = &NoopLogger{}
	}
	logger.Infof("Fetching all activities from Garmin Connect")

	if c.garthClient == nil {
		if err := c.Connect(logger); err != nil {
			return nil, err
		}
	}

	var allActivities []models.Activity
	pageSize := 100
	start := 0
	timeout := 30 * time.Second

	for {
		logger.Infof("Fetching activities from offset %d", start)

		// Create context with timeout for this page
		pageCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		resp, err := c.cb.Execute(func() (interface{}, error) {
			return c.garthClient.GetActivities(pageSize, start)
		})
		if err != nil {
			if pageCtx.Err() == context.DeadlineExceeded {
				logger.Errorf("Timeout fetching activities from offset %d", start)
			} else {
				logger.Errorf("Failed to fetch activities (circuit breaker): %v", err)
			}
			return nil, err
		}
		garthActivities := resp.([]*garth.Activity)

		if len(garthActivities) == 0 {
			break
		}

		logger.Infof("Fetched %d activities from offset %d", len(garthActivities), start)

		for _, ga := range garthActivities {
			// Skip activities with invalid start time
			if ga.StartTimeGMT.IsZero() {
				logger.Warnf("Activity %d has invalid start time", ga.ActivityID)
				continue
			}

			activity := models.Activity{
				ID:        fmt.Sprintf("%d", ga.ActivityID),
				Name:      ga.ActivityName,
				Type:      ga.ActivityType.TypeKey,
				Date:      ga.StartTimeGMT.Time,
				Distance:  ga.Distance,
				Duration:  time.Duration(ga.Duration) * time.Second,
				Elevation: ga.ElevationGain,
				Calories:  int(ga.Calories),
			}

			// Populate metrics
			if ga.AverageHR > 0 {
				activity.Metrics.AvgHeartRate = int(ga.AverageHR)
			}
			if ga.MaxHR > 0 {
				activity.Metrics.MaxHeartRate = int(ga.MaxHR)
			}
			if ga.AverageSpeed > 0 {
				activity.Metrics.AvgSpeed = ga.AverageSpeed * 3.6
			}
			if ga.ElevationGain > 0 {
				activity.Metrics.ElevationGain = ga.ElevationGain
			}
			if ga.ElevationLoss > 0 {
				activity.Metrics.ElevationLoss = ga.ElevationLoss
			}
			if ga.AverageSpeed > 0 && ga.Distance > 0 {
				activity.Metrics.AvgPace = (ga.Duration / ga.Distance) * 1000
			}

			allActivities = append(allActivities, activity)
		}

		// Break if we got fewer than page size
		if len(garthActivities) < pageSize {
			break
		}

		start += len(garthActivities)

		// Increase timeout for next page in case of large datasets
		timeout += 10 * time.Second
	}

	logger.Infof("Successfully fetched %d activities in total", len(allActivities))
	return allActivities, nil
}

func (c *Client) DownloadActivityFile(ctx context.Context, activityID string, format string, logger Logger) ([]byte, error) {
	if logger == nil {
		logger = &NoopLogger{}
	}
	logger.Infof("Downloading %s file for activity %s", format, activityID)

	if c.garthClient == nil {
		if err := c.Connect(logger); err != nil {
			return nil, err
		}
	}

	// Construct download URL based on format
	var path string
	switch format {
	case "gpx":
		path = fmt.Sprintf("/download-service/export/gpx/activity/%s", activityID)
	case "tcx":
		path = fmt.Sprintf("/download-service/export/tcx/activity/%s", activityID)
	case "fit":
		path = fmt.Sprintf("/download-service/files/activity/%s", activityID)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", format)
	}

	// Wrap download with circuit breaker
	resp, err := c.cb.Execute(func() (interface{}, error) {
		return c.garthClient.Download(path)
	})
	if err != nil {
		logger.Errorf("Failed to download %s file for activity %s (circuit breaker): %v", format, activityID, err)
		return nil, err
	}
	data := resp.([]byte)

	logger.Infof("Successfully downloaded %s file for activity %s (%d bytes)", format, activityID, len(data))
	return data, nil
}
