package timeseriesgo

import (
	"errors"
	"time"
)

type MeanAndVariance struct {
	mean               float64
	sampleVariance     float64
	populationVariance float64
}

/**
 * Calculates the mean and variance of the values in the TimeSeries.
 *
 * @return A MeanAndVariance struct containing the mean and variance, or an error if the TimeSeries is empty.
 */
func (ts *TimeSeries) GetMeanAndVariance() (MeanAndVariance, error) {
	if ts.IsEmpty() {
		return MeanAndVariance{}, errors.New("TimeSeries is empty")
	} else {
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
			mean:               mean,
			sampleVariance:     sampleVariance,
			populationVariance: populationVariance,
		}, nil
	}
}

func (ts *TimeSeries) MovingAverage(window time.Duration) TimeSeries {
	if ts.IsEmpty() {
		return Empty()
	}

	if window <= 0 {
		cloned := make([]DataPoint, len(ts.datapoints))
		copy(cloned, ts.datapoints)
		return TimeSeries{datapoints: cloned}
	}

	result := Empty()
	left := 0
	runningSum := 0.0

	for right, dp := range ts.datapoints {
		runningSum += dp.value

		// Maintain window (t-window, t] to match RollingWindow semantics.
		for left <= right && dp.timestamp.Sub(ts.datapoints[left].timestamp) >= window {
			runningSum -= ts.datapoints[left].value
			left++
		}

		count := right - left + 1
		result.AddPoint(DataPoint{
			timestamp: dp.timestamp,
			value:     runningSum / float64(count),
		})
	}

	return result
}
