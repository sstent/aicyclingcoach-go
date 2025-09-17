package garmin

import (
	"strconv"
	"time"

	"github.com/sstent/fitness-tui/internal/garmin/garth/types" // Use internal types
	"github.com/sstent/fitness-tui/internal/tui/models"
)

// convertToInternalActivity converts Garmin activity to internal model
func convertToInternalActivity(activity *types.Activity) *models.Activity {
	return &models.Activity{
		ID:           strconv.FormatInt(activity.ActivityID, 10),
		Name:         activity.ActivityName,
		Description:  activity.Description,
		Type:         activity.ActivityType.TypeKey,
		ActivityType: activity.ActivityType.TypeKey,
		Date:         activity.StartTimeLocal.Time,
		Duration:     time.Duration(activity.Duration * float64(time.Second)),
		Distance:     activity.Distance,
		Elevation:    activity.ElevationGain,
		Calories:     int(activity.Calories),
		FilePath:     "",
		Metrics: models.ActivityMetrics{
			AvgHeartRate:  int(activity.AverageHR),
			MaxHeartRate:  int(activity.MaxHR),
			AvgSpeed:      activity.AverageSpeed,
			ElevationGain: activity.ElevationGain,
			ElevationLoss: activity.ElevationLoss,
			// Power and cadence metrics not available in Garmin API response
			AvgPower:   0,
			MaxPower:   0,
			AvgCadence: 0,
			MaxCadence: 0,
		},
	}
}
