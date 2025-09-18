package garmin

import (
	"context"

	"github.com/sstent/fitness-tui/internal/tui/models"
)

type MockClient struct {
	ConnectError       error
	Activities         []*models.Activity
	GetActivitiesError error
}

func (m *MockClient) Connect() error {
	return m.ConnectError
}

func (m *MockClient) GetActivities(ctx context.Context, limit int) ([]*models.Activity, error) {
	return m.Activities, m.GetActivitiesError
}
