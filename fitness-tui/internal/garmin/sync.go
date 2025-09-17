package garmin

import (
	"context"
	"fmt"
	"time"

	"github.com/sstent/fitness-tui/internal/storage"
)

// Sync performs the complete synchronization process
func (c *Client) Sync(ctx context.Context, storage *storage.ActivityStorage, logger Logger) (int, error) {
	// Create a context with timeout for the entire sync process
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Authenticate
	logger.Infof("Authenticating with Garmin Connect...")
	if err := c.Connect(logger); err != nil {
		logger.Errorf("Authentication failed: %v", err)
		if _, ok := err.(*AuthenticationError); ok {
			return 0, fmt.Errorf("authentication failed: please check your credentials and try again")
		}
		return 0, err
	}
	logger.Infof("Authentication successful")

	// Get all activities metadata
	logger.Infof("Fetching activity metadata...")
	activities, err := c.GetAllActivities(timeoutCtx, logger)
	if err != nil {
		logger.Errorf("Failed to fetch activities: %v", err)
		return 0, err
	}
	logger.Infof("Found %d activities", len(activities))

	// Download files for each activity
	downloadedFiles := 0
	for i := range activities {
		activity := &activities[i]
		// Check if context has been cancelled
		select {
		case <-timeoutCtx.Done():
			logger.Warnf("Sync cancelled due to timeout")
			return downloadedFiles, timeoutCtx.Err()
		default:
		}

		logger.Infof("Processing activity %d/%d: %s", i+1, len(activities), activity.Name)

		// Only download if file doesn't exist
		if activity.FilePath == "" {
			logger.Infof("File missing for activity %s, attempting download...", activity.ID)
			var data []byte
			var format string
			var err error

			// First try FIT (preferred)
			logger.Infof("Trying FIT download for %s...", activity.ID)
			data, err = c.DownloadActivityFile(timeoutCtx, activity.ID, "fit", logger)
			if err == nil {
				format = "fit"
				logger.Infof("FIT download successful for %s (%d bytes)", activity.ID, len(data))
			} else {
				logger.Warnf("FIT download failed for %s: %v", activity.ID, err)

				// Fallback to GPX
				logger.Infof("Trying GPX download for %s...", activity.ID)
				data, err = c.DownloadActivityFile(timeoutCtx, activity.ID, "gpx", logger)
				if err == nil {
					format = "gpx"
					logger.Infof("GPX download successful for %s (%d bytes)", activity.ID, len(data))
				} else {
					logger.Warnf("GPX download failed for %s: %v", activity.ID, err)

					// Fallback to TCX
					logger.Infof("Trying TCX download for %s...", activity.ID)
					data, err = c.DownloadActivityFile(timeoutCtx, activity.ID, "tcx", logger)
					if err != nil {
						logger.Errorf("TCX download failed for %s: %v", activity.ID, err)
						continue
					}
					format = "tcx"
					logger.Infof("TCX download successful for %s (%d bytes)", activity.ID, len(data))
				}
			}

			// Save file to storage
			logger.Infof("Saving %s file for %s...", format, activity.ID)
			filePath, err := storage.SaveActivityFile(activity, data, format)
			if err != nil {
				logger.Errorf("Failed to save activity file for %s: %v", activity.ID, err)
				continue
			}
			logger.Infof("Saved file to %s", filePath)

			// Update activity with file path
			activity.FilePath = filePath
			downloadedFiles++
		} else {
			logger.Infof("File already exists for %s: %s", activity.ID, activity.FilePath)
		}

		// Save updated activity metadata
		logger.Infof("Saving metadata for %s...", activity.ID)
		if err := storage.Save(activity); err != nil {
			logger.Errorf("Failed to save activity metadata for %s: %v", activity.ID, err)
		}
	}

	return downloadedFiles, nil
}
