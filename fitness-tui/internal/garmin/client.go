package garmin

import (
	"context"
	"fmt"

	"github.com/sstent/fitness-tui/internal/tui/models"
	"github.com/sstent/go-garth"
)

type Client struct {
	client *garth.Client
	auth   *Auth
}

func NewClient(username, password, storagePath string) *Client {
	return &Client{
		client: garth.New(),
		auth:   NewAuth(username, password, storagePath),
	}
}

func (c *Client) GetActivities(ctx context.Context, limit int) ([]*models.Activity, error) {
	if err := c.auth.Connect(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	gActivities, err := c.client.GetActivities(ctx, 0, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch activities: %w", err)
	}

	activities := make([]*models.Activity, 0, len(gActivities))
	for _, ga := range gActivities {
		activities = append(activities, convertActivity(ga))
	}

	return activities, nil
}

func convertActivity(ga *garth.Activity) *models.Activity {
	return &models.Activity{
		ID:          ga.ID,
		Name:        ga.Name,
		Description: ga.Description,
		Type:        ga.Type,
		StartTime:   ga.StartTime,
		Distance:    ga.Distance,
		Duration:    ga.Duration,
		Elevation:   ga.Elevation,
		HeartRate:   ga.HeartRate,
	}
}
