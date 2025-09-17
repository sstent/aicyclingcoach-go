package garmin

import (
	"fmt"
	"time"

	"github.com/sstent/fitness-tui/internal/garmin/garth/types"
	"github.com/sstent/fitness-tui/internal/tui/models"
)

// MapGarminType converts Garmin activity types to internal analysis types
func MapGarminType(garminType types.ActivityType) string {
	switch garminType.TypeKey {
	case "cycling", "road_biking", "mountain_biking", "indoor_cycling":
		return "cycling"
	case "running", "treadmill_running", "trail_running":
		return "running"
	case "hiking", "walking":
		return "hiking"
	case "swimming":
		return "swimming"
	case "triathlon":
		return "triathlon"
	default:
		return "other"
	}
}

// ToActivityModel converts Garmin activity data to our internal Activity model
func ToActivityModel(garminActivity *types.Activity) *models.Activity {
	return &models.Activity{
		ID:           fmt.Sprintf("%d", garminActivity.ActivityID),
		Name:         garminActivity.ActivityName,
		Description:  garminActivity.Description,
		Type:         garminActivity.ActivityType.TypeKey,
		ActivityType: MapGarminType(garminActivity.ActivityType),
		Date:         garminActivity.StartTimeLocal.Time,
		Duration:     time.Duration(garminActivity.Duration * float64(time.Second)),
		Distance:     garminActivity.Distance,
		Elevation:    garminActivity.ElevationGain,
		Calories:     int(garminActivity.Calories),
		Metrics: models.ActivityMetrics{
			AvgHeartRate:  int(garminActivity.AverageHR),
			MaxHeartRate:  int(garminActivity.MaxHR),
			AvgSpeed:      garminActivity.AverageSpeed,
			ElevationGain: garminActivity.ElevationGain,
		},
	}
}
