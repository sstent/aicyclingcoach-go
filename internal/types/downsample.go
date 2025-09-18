package types

import "time"

type DownsampledPoint struct {
	Timestamp time.Time
	Value     float64
}

// DownsampleLTTB implements the Largest Triangle Three Buckets algorithm
// for time series downsampling. This is a corrected version that properly
// uses the current module path.
func DownsampleLTTB(data []float64, timestamps []time.Time, threshold int) []DownsampledPoint {
	if len(data) != len(timestamps) {
		panic("data and timestamps must be same length")
	}

	if threshold >= len(data) || threshold <= 0 {
		result := make([]DownsampledPoint, len(data))
		for i := range data {
			result[i] = DownsampledPoint{
				Timestamp: timestamps[i],
				Value:     data[i],
			}
		}
		return result
	}

	sampled := make([]DownsampledPoint, threshold)
	sampled[0] = DownsampledPoint{Timestamp: timestamps[0], Value: data[0]}

	bucketSize := float64(len(data)-2) / float64(threshold-2)
	a := 0

	for i := 0; i < threshold-2; i++ {
		avgRangeStart := int(float64(i+1)*bucketSize) + 1
		avgRangeEnd := int(float64(i+2)*bucketSize) + 1
		if avgRangeEnd > len(data) {
			avgRangeEnd = len(data)
		}

		var avgRange float64
		for j := avgRangeStart; j < avgRangeEnd; j++ {
			avgRange += data[j]
		}
		avgRange /= float64(avgRangeEnd - avgRangeStart)

		rangeOffs := int(float64(i)*bucketSize) + 1
		rangeTo := int(float64(i+1)*bucketSize) + 1
		if rangeTo > len(data) {
			rangeTo = len(data)
		}

		maxArea := -1.0
		nextAAt := 0
		for j := rangeOffs; j < rangeTo; j++ {
			area := areaSize(
				data[a],
				data[j],
				avgRange,
			)
			if area > maxArea {
				maxArea = area
				nextAAt = j
			}
		}

		sampled[i+1] = DownsampledPoint{
			Timestamp: timestamps[nextAAt],
			Value:     data[nextAAt],
		}
		a = nextAAt
	}

	sampled[threshold-1] = DownsampledPoint{
		Timestamp: timestamps[len(timestamps)-1],
		Value:     data[len(data)-1],
	}
	return sampled
}

func areaSize(a, b, avg float64) float64 {
	return (a-avg)*(a-avg) + (b-avg)*(b-avg)
}
