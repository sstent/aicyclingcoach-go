package models

import (
	"fmt"
	"time"
)

type Activity struct {
	ID       string
	Name     string
	Type     string
	Date     time.Time
	Duration time.Duration
	Distance float64 // meters
	Metrics  ActivityMetrics
}

type ActivityMetrics struct {
	AvgHeartRate  int
	MaxHeartRate  int
	AvgPace       float64 // seconds per km
	AvgSpeed      float64 // km/h
	ElevationGain float64 // meters
	ElevationLoss float64 // meters
}

func (a *Activity) FormattedDuration() string {
	hours := int(a.Duration.Hours())
	minutes := int(a.Duration.Minutes()) % 60
	return fmt.Sprintf("%02d:%02d", hours, minutes)
}

func (a *Activity) FormattedDistance() string {
	return fmt.Sprintf("%.2fkm", a.Distance/1000)
}

func (a *Activity) FormattedPace() string {
	minutes := int(a.Metrics.AvgPace) / 60
	seconds := int(a.Metrics.AvgPace) % 60
	return fmt.Sprintf("%d:%02d/km", minutes, seconds)
}
