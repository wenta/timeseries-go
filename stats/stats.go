package stats

import (
	"errors"
	"time"

	timeseriesgo "github.com/wenta/timeseries-go"
)

type MeanAndVariance struct {
	Mean               float64
	SampleVariance     float64
	PopulationVariance float64
}

/**
 * Calculates the mean and variance of the values in the TimeSeries.
 *
 * @ret
 **/
func GetMeanAndVariance(ts timeseriesgo.TimeSeries) (MeanAndVariance, error) {
	if ts.IsEmpty() {
		return MeanAndVariance{}, errors.New("TimeSeries is empty")
	}

	mean := ts.Sum() / float64(ts.Length())
	sampleVariance := 0.0
	for _, v := range ts.Values() {
		diff := v - mean
		sampleVariance += diff * diff
	}
	populationVariance := sampleVariance / float64(ts.Length())
	// Use sample variance (divide by n-1) to avoid underestimating stddev on small samples.
	if ts.Length() > 1 {
		sampleVariance /= float64(ts.Length() - 1)
	}
	return MeanAndVariance{
		Mean:               mean,
		SampleVariance:     sampleVariance,
		PopulationVariance: populationVariance,
	}, nil
}

// MovingAverage returns a rolling mean over the given time window (t-window, t].
// If window <= 0, it returns a shallow copy of the original series.
func MovingAverage(ts timeseriesgo.TimeSeries, window time.Duration) timeseriesgo.TimeSeries {
	if ts.IsEmpty() {
		return timeseriesgo.Empty()
	}

	if window <= 0 {
		cloned := ts.DataPoints()
		return timeseriesgo.FromDataPoints(cloned)
	}

	result := timeseriesgo.Empty()
	left := 0
	runningSum := 0.0
	points := ts.DataPoints()

	for right, dp := range points {
		runningSum += dp.Value

		// Maintain window (t-window, t] to match RollingWindow semantics.
		for left <= right && dp.Timestamp.Sub(points[left].Timestamp) >= window {
			runningSum -= points[left].Value
			left++
		}

		count := right - left + 1
		result.AddPoint(timeseriesgo.DataPoint{
			Timestamp: dp.Timestamp,
			Value:     runningSum / float64(count),
		})
	}

	return result
}

// MinMaxNormalize rescales values to the [0,1] range while preserving timestamps.
// Returns error if the TimeSeries is empty.
func MinMaxNormalize(ts timeseriesgo.TimeSeries) (timeseriesgo.TimeSeries, error) {
	if ts.IsEmpty() {
		return timeseriesgo.Empty(), errors.New("TimeSeries is empty")
	}

	points := ts.DataPoints()

	minVal := points[0].Value
	maxVal := points[0].Value

	// Find min and max
	for _, dp := range points {
		if dp.Value < minVal {
			minVal = dp.Value
		}
		if dp.Value > maxVal {
			maxVal = dp.Value
		}
	}

	result := timeseriesgo.Empty()

	// Avoid division by zero when all values are equal
	if maxVal == minVal {
		for _, dp := range points {
			result.AddPoint(timeseriesgo.DataPoint{
				Timestamp: dp.Timestamp,
				Value:     0.0,
			})
		}
		return result, nil
	}

	// Apply normalization
	denominator := maxVal - minVal
	for _, dp := range points {
		normalized := (dp.Value - minVal) / denominator
		result.AddPoint(timeseriesgo.DataPoint{
			Timestamp: dp.Timestamp,
			Value:     normalized,
		})
	}

	return result, nil
}
