package models

import (
	"fmt"
	"time"

	"github.com/sstent/fitness-tui/internal/types"
)

type Activity struct {
	ID           string
	Name         string
	Description  string
	Type         string // Garmin activity type (e.g., "running", "cycling")
	ActivityType string // Activity type for AI analysis prompts (e.g., "running", "cycling", "hiking")
	Date         time.Time
	Duration     time.Duration
	Distance     float64 // meters
	Elevation    float64
	Calories     int // in kilocalories
	Metrics      ActivityMetrics
	FilePath     string // Path to original activity file (e.g., GPX/FIT)
}

type ActivityMetrics struct {
	// Core metrics
	AvgHeartRate    int
	MaxHeartRate    int
	AvgPace         float64 // seconds per km
	AvgSpeed        float64 // km/h
	ElevationGain   float64 // meters
	ElevationLoss   float64 // meters
	RecoveryTime    int     // hours
	IntensityFactor float64

	// Raw data streams
	HeartRateData   []float64 `json:"heart_rate_data"`
	ElevationData   []float64 `json:"elevation_data"`
	PowerData       []float64 `json:"power_data,omitempty"`
	CadenceData     []float64 `json:"cadence_data,omitempty"`
	SpeedData       []float64 `json:"speed_data,omitempty"`
	TemperatureData []float64 `json:"temperature_data,omitempty"`

	// Downsampled metrics (for AI analysis)
	DownsampledHR      []types.DownsampledPoint `json:"downsampled_hr,omitempty"`
	DownsampledPower   []types.DownsampledPoint `json:"downsampled_power,omitempty"`
	DownsampledCadence []types.DownsampledPoint `json:"downsampled_cadence,omitempty"`
	DownsampledSpeed   []types.DownsampledPoint `json:"downsampled_speed,omitempty"`

	// Power metrics
	AvgPower        float64 `json:"avg_power"`
	MaxPower        float64 `json:"max_power"`
	NormalizedPower float64 `json:"normalized_power"`
	FTP             float64 `json:"ftp"` // Functional Threshold Power

	// Cadence metrics
	AvgCadence float64 `json:"avg_cadence"`
	MaxCadence float64 `json:"max_cadence"`

	// Running metrics
	GroundContactTime   int `json:"ground_contact_time"`  // milliseconds
	VerticalOscillation int `json:"vertical_oscillation"` // millimeters

	// Hiking metrics
	AscentRate     float64 `json:"ascent_rate"`     // meters/hour
	DescentRate    float64 `json:"descent_rate"`    // meters/hour
	AvgTemperature float64 `json:"avg_temperature"` // Celsius

	// Analysis-specific fields
	TargetZones         string  `json:"target_zones"`          // e.g., "Z2: 115-140bpm"
	TrainingLoad        float64 `json:"training_load"`         // Training load score
	TrainingStressScore float64 `json:"training_stress_score"` // TSS score
	ElevationProfile    string  `json:"elevation_profile"`     // e.g., "Hilly with 3 major climbs"
	FatigueLevel        string  `json:"fatigue_level"`         // e.g., "Moderate"
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
