package types

// Downsampler implements the Largest-Triangle-Three-Buckets algorithm
type Downsampler struct{}

// NewDownsampler creates a new Downsampler instance
func NewDownsampler() *Downsampler {
	return &Downsampler{}
}

// Process downsamples data using Largest-Triangle-Three-Buckets algorithm
func (d *Downsampler) Process(data []float64, threshold int) []float64 {
	if len(data) <= threshold || threshold <= 0 {
		return data
	}

	sampled := make([]float64, 0, threshold)
	sampled = append(sampled, data[0]) // First point

	bucketSize := float64(len(data)-2) / float64(threshold-2)

	for i := 1; i < threshold-1; i++ {
		bucketStart := int(float64(i-1)*bucketSize) + 1
		bucketEnd := int(float64(i)*bucketSize) + 1

		if bucketEnd >= len(data) {
			bucketEnd = len(data) - 1
		}

		maxArea := -1.0
		selectedPoint := data[bucketStart]

		for j := bucketStart; j < bucketEnd; j++ {
			area := triangleArea(
				data[bucketStart-1],
				data[j],
				data[bucketEnd],
			)

			if area > maxArea {
				maxArea = area
				selectedPoint = data[j]
			}
		}

		sampled = append(sampled, selectedPoint)
	}

	sampled = append(sampled, data[len(data)-1]) // Last point
	return sampled
}

// triangleArea calculates the area of a triangle formed by three points
func triangleArea(a, b, c float64) float64 {
	return abs((a-b)*(c-b)-(b-c)*(a-c)) / 2
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
