package models

import (
	"fmt"
	"time"
)

type Activity struct {
	ID          string
	Name        string
	Description string
	Type        string
	Date        time.Time
	Duration    time.Duration
	Distance    float64 // meters
	Elevation   float64
	Calories    int // in kilocalories
	Metrics     ActivityMetrics
	FilePath    string // Path to original activity file (e.g., GPX/FIT)
}

type ActivityMetrics struct {
	AvgHeartRate    int
	MaxHeartRate    int
	AvgPace         float64 // seconds per km
	AvgSpeed        float64 // km/h
	ElevationGain   float64 // meters
	ElevationLoss   float64 // meters
	TrainingStress  float64 // TSS score
	RecoveryTime    int     // hours
	IntensityFactor float64
	HeartRateData   []float64
	ElevationData   []float64
}

func (a *Activity) FormattedDuration() string {
	hours := int(a.Duration.Hours())
	minutes := int(a.Duration.Minutes()) % 60
	seconds := int(a.Duration.Seconds()) % 60
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func (a *Activity) FormattedDistance() string {
	return fmt.Sprintf("%.2fkm", a.Distance/1000)
}

func (a *Activity) FormattedPace() string {
	if a.Metrics.AvgPace <= 0 {
		return "--:--"
	}
	minutes := int(a.Metrics.AvgPace) / 60
	seconds := int(a.Metrics.AvgPace) % 60
	return fmt.Sprintf("%d:%02d/km", minutes, seconds)
}
