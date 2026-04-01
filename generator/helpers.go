package generator

import (
	"math"
	"math/rand/v2"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

func zipIndexValues(index []time.Time, values []float64) timeseriesgo.TimeSeries {
	ts, err := timeseriesgo.Zip(index, values)
	if err != nil {
		return timeseriesgo.Empty()
	}
	return ts
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func randomFloat64(rng *rand.Rand) float64 {
	if rng != nil {
		return rng.Float64()
	}
	return rand.Float64()
}

func randomIntN(rng *rand.Rand, n int) int {
	if rng != nil {
		return rng.IntN(n)
	}
	return rand.IntN(n)
}

func poissonCount(lambda float64, rng *rand.Rand) int {
	if lambda <= 0 {
		return 0
	}

	limit := math.Exp(-lambda)
	product := 1.0
	count := 0
	for product > limit {
		count++
		product *= randomFloat64(rng)
	}
	return count - 1
}
