package analysis

import (
	"time"
)

// DownsampledPoint represents a single data point in a downsampled metric series
type DownsampledPoint struct {
	TimeOffset int     `json:"time_offset"` // Seconds from activity start
	Value      float64 `json:"value"`       // Average value during this segment
	Min        float64 `json:"min,omitempty"`
	Max        float64 `json:"max,omitempty"`
}

// DownsampleMetric downsamples a metric array to the specified number of points
func DownsampleMetric(data []float64, duration time.Duration, targetPoints int) []DownsampledPoint {
	if len(data) == 0 || targetPoints <= 0 {
		return nil
	}

	totalSeconds := int(duration.Seconds())
	if totalSeconds <= 0 {
		return nil
	}

	// Calculate segment duration in seconds (floating point for accuracy)
	segmentDuration := float64(totalSeconds) / float64(targetPoints)
	if segmentDuration < 1 {
		segmentDuration = 1
	}

	// Preallocate segments array
	segments := make([]struct {
		sum   float64
		count int
		min   float64
		max   float64
	}, targetPoints)

	// Initialize min/max with zero values - we'll update them when we see data

	// Assign data points to segments
	for i, value := range data {
		// Calculate time offset for this data point
		timeOffset := float64(i) * float64(totalSeconds) / float64(len(data))
		segmentIndex := int(timeOffset / segmentDuration)

		// Clamp to valid segment range
		if segmentIndex >= targetPoints {
			segmentIndex = targetPoints - 1
		}

		seg := &segments[segmentIndex]
		seg.sum += value
		seg.count++

		// Initialize min/max on first value in segment
		if seg.count == 1 {
			seg.min = value
			seg.max = value
		} else {
			if value < seg.min {
				seg.min = value
			}
			if value > seg.max {
				seg.max = value
			}
		}
	}

	// Build results array
	results := make([]DownsampledPoint, targetPoints)
	for j := 0; j < targetPoints; j++ {
		seg := &segments[j]
		timeOffset := int(float64(j) * segmentDuration)

		if seg.count == 0 {
			// For empty segments, omit min/max
			results[j] = DownsampledPoint{
				TimeOffset: timeOffset,
				Value:      0,
			}
		} else {
			avg := seg.sum / float64(seg.count)
			results[j] = DownsampledPoint{
				TimeOffset: timeOffset,
				Value:      avg,
				Min:        seg.min,
				Max:        seg.max,
			}
		}
	}

	return results
}
