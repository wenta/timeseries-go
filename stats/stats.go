package stats

import (
	"errors"
	"math"
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
func Correlation(ts1, ts2 timeseriesgo.TimeSeries) (float64, error) {
	if ts1.IsEmpty() || ts2.IsEmpty() {
		return 0, errors.New("one or both TimeSeries are empty")
	}

	points1 := ts1.DataPoints()
	points2 := ts2.DataPoints()

	
	valueMap := make(map[int64]float64)
	for _, p := range points1 {
		valueMap[p.Timestamp.UnixNano()] = p.Value
	}

	var aligned1 []timeseriesgo.DataPoint
	var aligned2 []timeseriesgo.DataPoint

	for _, p := range points2 {
		if v, ok := valueMap[p.Timestamp.UnixNano()]; ok {
			aligned1 = append(aligned1, timeseriesgo.DataPoint{
				Timestamp: p.Timestamp,
				Value:     v,
			})
			aligned2 = append(aligned2, p)
		}
	}

	if len(aligned1) < 2 {
		return 0, errors.New("not enough aligned points")
	}

	series1 := timeseriesgo.FromDataPoints(aligned1)
	series2 := timeseriesgo.FromDataPoints(aligned2)

	stats1, err := GetMeanAndVariance(series1)
	if err != nil {
		return 0, err
	}

	stats2, err := GetMeanAndVariance(series2)
	if err != nil {
		return 0, err
	}

	
	if stats1.SampleVariance == 0 || stats2.SampleVariance == 0 {
		return 0, errors.New("zero variance in one of the series")
	}

	
	covariance := 0.0
	for i := range aligned1 {
		dx := aligned1[i].Value - stats1.Mean
		dy := aligned2[i].Value - stats2.Mean
		covariance += dx * dy
	}

	covariance /= float64(len(aligned1) - 1)

	
	correlation := covariance / (math.Sqrt(stats1.SampleVariance) * math.Sqrt(stats2.SampleVariance))

	return correlation, nil
}
