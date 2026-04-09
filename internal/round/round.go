package round

import "math"

// Float32 rounds like Java UpdateScheduler.roundValue: scaled float, math.Round, then divide.
func Float32(value float32, precision int) float32 {
	if precision < 0 {
		precision = 0
	}
	scale := math.Pow(10, float64(precision))
	temp := float64(value) * scale
	return float32(math.Round(temp) / scale)
}
